package payment

import (
	"bot/src/online/db"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Prodamus docs: https://help.prodamus.ru/payform/integracii/rest-api/instrukcii-dlya-samostoyatelnaya-integracii-servisov
//
// Env:
//   PRODAMUS_FORM_URL        https://<domain>.payform.ru
//   PRODAMUS_SECRET          secret key from the form settings
//   PRODAMUS_SUBSCRIPTION_ID id from <domain>.payform.ru/settings/subscriptions/
//   PRODAMUS_SUCCESS_URL     optional redirect after payment
//
// Webhook URL to set in Prodamus (form settings and the subscriptions tab):
//   <PUBLIC_BASE_URL>/api/online/prodamus/webhook

const orderPrefix = "tg-"

type prodamus struct {
	formURL        string
	secret         string
	subscriptionID string
	successURL     string
}

func newProdamus() (*prodamus, bool) {
	p := &prodamus{
		formURL:        strings.TrimRight(os.Getenv("PRODAMUS_FORM_URL"), "/"),
		secret:         os.Getenv("PRODAMUS_SECRET"),
		subscriptionID: os.Getenv("PRODAMUS_SUBSCRIPTION_ID"),
		successURL:     os.Getenv("PRODAMUS_SUCCESS_URL"),
	}
	return p, p.formURL != "" && p.secret != "" && p.subscriptionID != ""
}

func (p *prodamus) Name() db.OnlinePaymentProvider {
	return db.OnlinePaymentProviderProdamus
}

func (p *prodamus) Button() string {
	return "Карта РФ или СБП 💳"
}

// PaymentLink builds a signed payform link. The Telegram ID goes into order_id,
// the signature keeps it from being changed, Prodamus sends it back as order_num.
func (p *prodamus) PaymentLink(userId int64) (string, error) {
	params := map[string]any{
		"do":           "pay",
		"order_id":     fmt.Sprintf("%s%d-%d", orderPrefix, userId, time.Now().Unix()),
		"subscription": p.subscriptionID,
	}
	if p.successURL != "" {
		params["urlSuccess"] = p.successURL
	}

	signature, err := prodamusSign(params, p.secret)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	for k, v := range params {
		query.Set(k, v.(string))
	}
	query.Set("signature", signature)

	return p.formURL + "/?" + query.Encode(), nil
}

func (p *prodamus) ParseWebhook(r *http.Request) (Event, error) {
	if err := r.ParseForm(); err != nil {
		return Event{}, err
	}
	data := parsePHPForm(r.PostForm)
	log.Printf("prodamus webhook: %v", data)

	expected, err := prodamusSign(data, p.secret)
	if err != nil {
		return Event{}, err
	}
	if !hmac.Equal([]byte(expected), []byte(strings.ToLower(r.Header.Get("Sign")))) {
		return Event{}, ErrBadSignature
	}

	subscription, _ := data["subscription"].(map[string]any)
	userId := parseOrderNum(str(data["order_num"]))

	// The webhook URL gets every payment of the form, not only the club ones.
	if userId == 0 && str(subscription["id"]) != p.subscriptionID {
		return Event{}, ErrIgnored
	}

	status := str(data["payment_status"])
	if desc := str(data["payment_status_description"]); desc != "" {
		status += " (" + desc + ")"
	}

	return Event{
		Paid:        str(data["payment_status"]) == "success",
		PaymentRef:  str(data["order_id"]),
		CustomerRef: str(subscription["profile_id"]),
		UserID:      userId,
		Amount:      str(data["sum"]),
		Currency:    str(data["currency"]),
		Status:      status,
	}, nil
}

// parseOrderNum takes the Telegram ID out of "tg-<id>-<unix>".
func parseOrderNum(orderNum string) int64 {
	rest, ok := strings.CutPrefix(orderNum, orderPrefix)
	if !ok {
		return 0
	}
	idStr, _, _ := strings.Cut(rest, "-")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

// prodamusSign mirrors Hmac::create from the Prodamus PHP library:
// json_encode($data, JSON_UNESCAPED_UNICODE) of string values with sorted keys,
// so "/" is escaped as "\/" and unicode is kept as is.
func prodamusSign(data map[string]any, secret string) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	// json.Encoder sorts map keys.
	if err := enc.Encode(data); err != nil {
		return "", err
	}
	payload := strings.ReplaceAll(strings.TrimSuffix(buf.String(), "\n"), "/", `\/`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// parsePHPForm turns "products[0][name]=x" style keys into nested maps the way
// PHP fills $_POST. Maps with keys 0..n-1 become lists, like PHP json_encode does.
func parsePHPForm(form url.Values) map[string]any {
	root := map[string]any{}

	for key, values := range form {
		path := splitPHPKey(key)
		node := root
		for _, k := range path[:len(path)-1] {
			child, ok := node[k].(map[string]any)
			if !ok {
				child = map[string]any{}
				node[k] = child
			}
			node = child
		}
		node[path[len(path)-1]] = values[len(values)-1]
	}

	for k, v := range root {
		root[k] = toLists(v)
	}
	return root
}

func splitPHPKey(key string) []string {
	name, rest, ok := strings.Cut(key, "[")
	if !ok || !strings.HasSuffix(rest, "]") {
		return []string{key}
	}
	return append([]string{name}, strings.Split(strings.TrimSuffix(rest, "]"), "][")...)
}

func toLists(v any) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}

	for k, child := range m {
		m[k] = toLists(child)
	}

	keys := make([]int, 0, len(m))
	for k := range m {
		i, err := strconv.Atoi(k)
		if err != nil || strconv.Itoa(i) != k {
			return m
		}
		keys = append(keys, i)
	}
	sort.Ints(keys)
	for i, k := range keys {
		if i != k {
			return m
		}
	}

	list := make([]any, len(keys))
	for _, k := range keys {
		list[k] = m[strconv.Itoa(k)]
	}
	return list
}
