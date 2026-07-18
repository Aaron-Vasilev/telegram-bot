package payment

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/controller"
	"bot/src/db"
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

		userName := telegramUserID
		userEmail := ""
		if uid, err := strconv.ParseInt(telegramUserID, 10, 64); err == nil {
			if user, err := db.Query.GetUserWithMembership(b.Ctx, uid); err == nil && user.Name != "" {
				userName = user.Name
			}
		}

		approveURL, err := createMorningPaymentForm(morningFormRequest{
			description: plan.Label,
			amount:      plan.Price,
			clientName:  userName,
			clientEmail: userEmail,
			custom:      telegramUserID + ":" + planKey,
			successPath: "/api/payment-success",
			failurePath: "/yoga/?failed=1",
			saveToken:   false,
		})
		if err != nil {
			b.Error("morning create form: " + err.Error())
			http.Error(w, "failed to create payment", http.StatusInternalServerError)
			return
		}

		b.SendHTML(common.AdminChatID(), fmt.Sprintf(
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
		paymentID := r.URL.Query().Get("id")
		if paymentID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		payment, err := fetchMorningPayment(paymentID)
		if err != nil {
			b.Error("morning fetch payment: " + err.Error())
			http.Error(w, "verify failed", http.StatusInternalServerError)
			return
		}

		if payment.Status != morningPaymentSuccess {
			b.Error(fmt.Sprintf("payment-success: unexpected status %d for %s", payment.Status, paymentID))
			http.Error(w, "payment not successful", http.StatusBadRequest)
			return
		}

		parts := strings.SplitN(payment.Custom, ":", 2)
		if len(parts) != 2 {
			b.Error("payment-success: malformed custom " + payment.Custom)
			http.Error(w, "malformed custom", http.StatusBadRequest)
			return
		}
		telegramUserID, planKey := parts[0], parts[1]

		if uid, err := strconv.ParseInt(telegramUserID, 10, 64); err == nil {
			switch planKey {
			case "membership_1":
				controller.UpdateMembership(b.Ctx, uid, 1)
			case "membership_2":
				controller.UpdateMembership(b.Ctx, uid, 2)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;padding:60px;margin-top:50%">
<h1>✅ Payment successful!</h1><p>See you on the mat 🧘</p></body></html>`)

		notifyOneTimeCapture(b, payment, telegramUserID, planKey)
	}
}

func notifyOneTimeCapture(b *bot.Bot, p morningPayment, telegramUserID, planKey string) {
	plan := plans[planKey]
	msg := fmt.Sprintf(
		"✅ Payment received!\n\nCustomer: <b>%s</b>\nTelegram ID: <code>%s</code>\nEmail: %s\nPlan: <b>%s</b>\nAmount: <b>%.2f ILS</b>\nPayment: <code>%s</code>",
		p.ClientName, telegramUserID, p.ClientEmail, plan.Label, p.Amount, p.ID,
	)
	b.SendHTML(common.AdminChatID(), msg)
}

// ── shared Morning helper types (used by both flows) ──

type morningFormRequest struct {
	description string
	amount      float64
	clientName  string
	clientEmail string
	custom      string
	successPath string
	failurePath string
	saveToken   bool
}

const morningPaymentSuccess = 1 // Morning status code for completed payment

type morningPayment struct {
	ID          string  `json:"id"`
	Status      int     `json:"status"`
	Amount      float64 `json:"amount"`
	Custom      string  `json:"custom"`
	ClientName  string  `json:"clientName"`
	ClientEmail string  `json:"clientEmail"`
	CardToken   string  `json:"cardToken"`
}

func createMorningPaymentForm(req morningFormRequest) (string, error) {
	token, err := getMorningToken()
	if err != nil {
		return "", err
	}

	base := os.Getenv("PUBLIC_BASE_URL")

	client := map[string]any{"name": req.clientName}
	if req.clientEmail != "" {
		client["emails"] = []string{req.clientEmail}
	}

	body := map[string]any{
		"description": req.description,
		"type":        400,
		"lang":        "he",
		"currency":    "ILS",
		"vatType":     0,
		"amount":      req.amount,
		"maxPayments": 1,
		"pluginId":    os.Getenv("MORNING_PLUGIN_ID"),
		"client":      client,
		"successUrl":  base + req.successPath,
		"failureUrl":  base + req.failurePath,
		"notifyUrl":   base + "/api/morning-webhook",
		"custom":      req.custom,
	}
	if req.saveToken {
		body["saveMe"] = true
	}

	data, _ := json.Marshal(body)
	httpReq, _ := http.NewRequest(http.MethodPost, morningBase()+"/payments/form", bytes.NewBuffer(data))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("morning form: status %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		URL      string `json:"url"`
		FormLink string `json:"formLink"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.URL != "" {
		return result.URL, nil
	}
	if result.FormLink != "" {
		return result.FormLink, nil
	}
	return "", fmt.Errorf("morning: no url in form response")
}

func fetchMorningPayment(paymentID string) (morningPayment, error) {
	token, err := getMorningToken()
	if err != nil {
		return morningPayment{}, err
	}

	req, _ := http.NewRequest(http.MethodGet, morningBase()+"/payments/"+paymentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return morningPayment{}, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	log.Printf("morning payment fetch: %s", raw)

	if resp.StatusCode >= 300 {
		return morningPayment{}, fmt.Errorf("morning fetch: status %d: %s", resp.StatusCode, raw)
	}

	var p morningPayment
	if err := json.Unmarshal(raw, &p); err != nil {
		return morningPayment{}, err
	}
	return p, nil
}
