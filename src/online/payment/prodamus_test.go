package payment

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Expected signatures are computed outside Go, the way the Prodamus PHP library does it.

func TestProdamusSignLink(tt *testing.T) {
	params := map[string]any{
		"do":           "pay",
		"order_id":     "tg-42-1700000000",
		"subscription": "123",
		"urlSuccess":   "https://t.me/yoga_bot",
	}

	got, err := prodamusSign(params, "secret")
	if err != nil {
		tt.Fatal(err)
	}
	if want := "3e1cd5eb8dde16f76c30371dad37c92cad334c8201c24a25872769a8b9179e3c"; got != want {
		tt.Errorf("got %s, want %s", got, want)
	}
}

func TestProdamusParseWebhook(tt *testing.T) {
	form := url.Values{
		"order_id":                 {"777"},
		"order_num":                {"tg-42-1700000000"},
		"payment_status":           {"success"},
		"products[0][name]":        {"Клуб йоги"},
		"products[0][price]":       {"4200.00"},
		"subscription[id]":         {"123"},
		"subscription[profile_id]": {"555"},
		"sum":                      {"4200.00"},
	}
	p := &prodamus{secret: "secret", subscriptionID: "123"}

	newRequest := func(sign string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Sign", sign)
		return r
	}

	event, err := p.ParseWebhook(newRequest("9ff8402e8480d2eac32735b883ec4d7bda91bed891663993b8a36f20e493b0dc"))
	if err != nil {
		tt.Fatal(err)
	}
	want := Event{
		Paid:        true,
		PaymentRef:  "777",
		CustomerRef: "555",
		UserID:      42,
		Amount:      "4200.00",
		Status:      "success",
	}
	if event != want {
		tt.Errorf("got %+v, want %+v", event, want)
	}

	if _, err := p.ParseWebhook(newRequest("deadbeef")); err != ErrBadSignature {
		tt.Errorf("got %v, want ErrBadSignature", err)
	}
}

func TestParseOrderNum(tt *testing.T) {
	tests := map[string]int64{
		"tg-362575139-1700000000": 362575139,
		"tg-42":                   42,
		"tg-":                     0,
		"order-42":                0,
		"":                        0,
	}
	for in, want := range tests {
		if got := parseOrderNum(in); got != want {
			tt.Errorf("parseOrderNum(%q) = %d, want %d", in, got, want)
		}
	}
}
