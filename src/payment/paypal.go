package payment

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/db"
	"bot/src/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func StartPaymentServer(b *bot.Bot) {
	b.Mux.HandleFunc("/api/create-payment", createPaymentHandler(b))
	b.Mux.HandleFunc("/api/payment-success", paymentSuccessHandler(b))
	b.Mux.HandleFunc("/api/paypal", paypalWebhookHandler(b))

	b.Mux.Handle("/yoga/", http.StripPrefix("/yoga", http.FileServer(http.Dir("pages/yoga"))))
}

type planInfo struct {
	Label string
	Price float64
}

var plans = map[string]planInfo{
	"membership_1": {"One lesson/week — 4 week membership", 280},
	"membership_2": {"Two lessons/week — 4 week membership", 400},
	"single_first": {"First yoga class", 70},
	"single":       {"Drop-in class", 90},
}

func paypalBase() string {
	if os.Getenv("ENV") == "production" {
		return "https://api-m.paypal.com"
	}
	return "https://api-m.sandbox.paypal.com"
}

func createPaymentHandler(b *bot.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		telegramUserID := r.FormValue("telegram_user_id")
		planKey := r.FormValue("plan")

		plan, ok := plans[planKey]
		if !ok || telegramUserID == "" {
			http.Error(w, "missing telegram_user_id or plan", http.StatusBadRequest)
			return
		}

		approveURL, err := createOrder(plan, planKey, telegramUserID)
		if err != nil {
			b.Error("paypal create order: " + err.Error())
			http.Error(w, "failed to create payment", http.StatusInternalServerError)
			return
		}

		userName := telegramUserID
		if uid, err := strconv.ParseInt(telegramUserID, 10, 64); err == nil {
			if user, err := db.Query.GetUserWithMembership(b.Ctx, uid); err == nil && user.Name != "" {
				userName = user.Name
			}
		}

		b.SendHTML(receiverID(), fmt.Sprintf(
			"🛒 Checkout started!\n\n%s\nPlan: %s",
			userName, plan.Label,
		))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"redirect_url": approveURL,
		})
	}
}

func paymentSuccessHandler(b *bot.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("token")
		if orderID == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}

		capture, err := captureOrder(orderID)
		if err != nil {
			b.Error("paypal capture: " + err.Error())
			http.Error(w, "capture failed", http.StatusInternalServerError)
			return
		}

		if uid, err := strconv.ParseInt(capture.TelegramUserID, 10, 64); err == nil {
			switch capture.PlanKey {
			case "membership_1":
				controller.UpdateMembership(b.Ctx, uid, 1)
			case "membership_2":
				controller.UpdateMembership(b.Ctx, uid, 2)
			}
		}

		if err := createMorningReceipt(capture); err != nil {
			b.Error("morning receipt: " + err.Error())
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;padding:60px">
<h1>✅ Payment successful!</h1><p>See you on the mat 🧘</p></body></html>`)

		notifyCapture(b, capture)
	}
}

// paypalWebhookHandler handles POST /paypal — called by PayPal for async events.
type paypalEvent struct {
	EventType string         `json:"event_type"`
	Resource  map[string]any `json:"resource"`
}

func paypalWebhookHandler(b *bot.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusInternalServerError)
			return
		}

		if !verifyPaypalWebhook(r, body) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var event paypalEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		switch event.EventType {
		case "PAYMENT.CAPTURE.DENIED":
			name := extractPayerName(event.Resource)
			amount := extractAmount(event.Resource)
			b.SendHTML(receiverID(), fmt.Sprintf("❌ Payment denied!\n\nCustomer: <b>%s</b>\nAmount: <b>%s</b>", name, amount))
		}

		w.WriteHeader(http.StatusOK)
	}
}

// ── PayPal API calls ──────────────────────────────────────────────────────────

type captureResult struct {
	ID             string
	Name           string
	Email          string
	Amount         string
	TelegramUserID string
	PlanKey        string
}

func createOrder(plan planInfo, planKey, telegramUserID string) (string, error) {
	token, err := getPaypalToken()
	if err != nil {
		return "", err
	}

	base := os.Getenv("PAYPAL_WEBHOOK_URL")
	payload := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{{
			"description": plan.Label,
			"custom_id":   telegramUserID + ":" + planKey,
			"amount": map[string]any{
				"currency_code": "ILS",
				"value":         fmt.Sprintf("%.2f", plan.Price),
			},
		}},
		"application_context": map[string]any{
			"brand_name":  "Yoga Studio",
			"user_action": "PAY_NOW",
			"return_url":  base + "/api/payment-success",
			"cancel_url":  base + "/api/payment-cancel",
		},
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, paypalBase()+"/v2/checkout/orders", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	for _, link := range result.Links {
		if link.Rel == "approve" {
			return link.Href, nil
		}
	}
	return "", fmt.Errorf("no approve link in PayPal response")
}

func captureOrder(orderID string) (captureResult, error) {
	token, err := getPaypalToken()
	if err != nil {
		return captureResult{}, err
	}

	req, _ := http.NewRequest(http.MethodPost,
		paypalBase()+"/v2/checkout/orders/"+orderID+"/capture",
		bytes.NewBufferString("{}"),
	)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return captureResult{}, err
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	log.Printf("paypal capture raw: %s", rawBody)

	var body map[string]any
	json.Unmarshal(rawBody, &body)

	result := captureResult{ID: stringField(body, "id")}

	if payer, ok := body["payer"].(map[string]any); ok {
		if name, ok := payer["name"].(map[string]any); ok {
			result.Name = stringField(name, "given_name") + " " + stringField(name, "surname")
		}
		result.Email = stringField(payer, "email_address")
	}

	if units, ok := body["purchase_units"].([]any); ok && len(units) > 0 {
		if unit, ok := units[0].(map[string]any); ok {
			if payments, ok := unit["payments"].(map[string]any); ok {
				if captures, ok := payments["captures"].([]any); ok && len(captures) > 0 {
					if capture, ok := captures[0].(map[string]any); ok {
						result.Amount = extractAmount(capture)
						parts := strings.SplitN(stringField(capture, "custom_id"), ":", 2)
						result.TelegramUserID = parts[0]
						if len(parts) == 2 {
							result.PlanKey = parts[1]
						}
					}
				}
			}
		}
	}

	return result, nil
}

func notifyCapture(b *bot.Bot, c captureResult) {
	msg := fmt.Sprintf(
		"✅ Payment received!\n\nCustomer: <b>%s</b>\nTelegram ID: <code>%s</code>\nEmail: %s\nAmount: <b>%s</b>\nOrder: <code>%s</code>",
		c.Name, c.TelegramUserID, c.Email, c.Amount, c.ID,
	)
	b.SendHTML(receiverID(), msg)
}

func extractAmount(resource map[string]any) string {
	if amt, ok := resource["amount"].(map[string]any); ok {
		v := stringField(amt, "value")
		cur := stringField(amt, "currency_code")
		if v != "" {
			return v + " " + cur
		}
	}
	return "unknown"
}

func extractPayerName(resource map[string]any) string {
	if payer, ok := resource["payer"].(map[string]any); ok {
		if name, ok := payer["name"].(map[string]any); ok {
			g := stringField(name, "given_name")
			s := stringField(name, "surname")
			if g != "" {
				return g + " " + s
			}
		}
	}
	return "unknown"
}

func verifyPaypalWebhook(r *http.Request, body []byte) bool {
	token, err := getPaypalToken()
	if err != nil {
		return false
	}

	payload := map[string]any{
		"auth_algo":         r.Header.Get("PAYPAL-AUTH-ALGO"),
		"cert_url":          r.Header.Get("PAYPAL-CERT-URL"),
		"transmission_id":   r.Header.Get("PAYPAL-TRANSMISSION-ID"),
		"transmission_sig":  r.Header.Get("PAYPAL-TRANSMISSION-SIG"),
		"transmission_time": r.Header.Get("PAYPAL-TRANSMISSION-TIME"),
		"webhook_id":        os.Getenv("PAYPAL_WEBHOOK_ID"),
		"webhook_event":     json.RawMessage(body),
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost,
		paypalBase()+"/v1/notifications/verify-webhook-signature",
		bytes.NewBuffer(data),
	)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result struct {
		VerificationStatus string `json:"verification_status"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.VerificationStatus == "SUCCESS"
}

func getPaypalToken() (string, error) {
	req, _ := http.NewRequest(http.MethodPost,
		paypalBase()+"/v1/oauth2/token",
		bytes.NewBufferString("grant_type=client_credentials"),
	)
	req.SetBasicAuth(os.Getenv("PAYPAL_CLIENT_ID"), os.Getenv("PAYPAL_SECRET"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty token")
	}
	return result.AccessToken, nil
}

func receiverID() int64 {
	if os.Getenv("ENV") == "production" {
		return utils.VIOLETTA_ID
	}

	return utils.MY_ID
}
