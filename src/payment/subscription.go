package payment

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/online/db"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const subscriptionPriceILS = 111.0

func StartSubscriptionServer(b *bot.Bot) {
	b.Mux.HandleFunc("/api/create-subscription", createSubscriptionHandler(b))
	b.Mux.HandleFunc("/api/subscription-success", subscriptionSuccessHandler(b))
	b.Mux.HandleFunc("/api/morning-webhook", morningWebhookHandler(b))

	b.Mux.Handle("/online/", http.StripPrefix("/online", http.FileServer(http.Dir("pages/online"))))
}

func createSubscriptionHandler(b *bot.Bot) http.HandlerFunc {
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
		if telegramUserID == "" {
			http.Error(w, "missing telegram_user_id", http.StatusBadRequest)
			return
		}

		tgUserId, err := strconv.ParseInt(telegramUserID, 10, 64)
		if err != nil {
			http.Error(w, "invalid telegram_user_id", http.StatusBadRequest)
			return
		}

		userName := r.FormValue("customer_name")
		if userName == "" {
			userName = telegramUserID
			if user, err := db.Query.GetUser(b.Ctx, tgUserId); err == nil {
				userName = strings.TrimSpace(user.FirstName + " " + user.LastName)
			}
		}

		url, err := createMorningPaymentForm(morningFormRequest{
			description: "Yoletta Online — monthly subscription",
			amount:      subscriptionPriceILS,
			clientName:  userName,
			custom:      telegramUserID,
			successPath: "/api/subscription-success",
			failurePath: "/online/?failed=1",
			saveToken:   true,
		})
		if err != nil {
			b.Error("morning create subscription form: " + err.Error())
			http.Error(w, "failed to create subscription", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"redirect_url": url})
	}
}

func subscriptionSuccessHandler(b *bot.Bot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paymentID := r.URL.Query().Get("id")
		if paymentID == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		payment, err := fetchMorningPayment(paymentID)
		if err != nil {
			b.Error("morning fetch subscription: " + err.Error())
			http.Error(w, "verify failed", http.StatusInternalServerError)
			return
		}

		if payment.Status != morningPaymentSuccess {
			b.Error(fmt.Sprintf("subscription-success: unexpected status %d for %s", payment.Status, paymentID))
			http.Error(w, "payment not successful", http.StatusBadRequest)
			return
		}

		tgUserId, err := strconv.ParseInt(payment.Custom, 10, 64)
		if err != nil {
			b.Error("subscription-success: bad custom " + payment.Custom)
			http.Error(w, "malformed custom", http.StatusBadRequest)
			return
		}

		if err := activateSubscription(b, tgUserId, payment.ID, payment.CardToken); err != nil {
			b.Error("activateSubscription: " + err.Error())
			http.Error(w, "activation failed", http.StatusInternalServerError)
			return
		}

		b.SendHTML(common.AdminChatID(), fmt.Sprintf(
			"✅ Online subscription activated!\n%s\nMorning payment: <code>%s</code>",
			formatUser(b, tgUserId), payment.ID,
		))

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;padding:60px;margin-top:50%">
<h1>✅ Subscription activated!</h1><p>See you on the mat 🧘</p></body></html>`)
	}
}

type morningEvent struct {
	Type    string         `json:"type"`
	Payment morningPayment `json:"payment"`
}

func morningWebhookHandler(b *bot.Bot) http.HandlerFunc {
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
		log.Printf("morning webhook: %s", body)

		var event morningEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if event.Payment.Status != morningPaymentSuccess && event.Payment.Status != 0 {
			b.SendHTML(common.AdminChatID(), fmt.Sprintf(
				"❌ Morning payment failed: <code>%s</code> status=%d",
				event.Payment.ID, event.Payment.Status,
			))
		}

		w.WriteHeader(http.StatusOK)
	}
}

func formatUser(b *bot.Bot, userID int64) string {
	user, err := db.Query.GetUser(b.Ctx, userID)
	if err != nil {
		return fmt.Sprintf("Telegram ID: <code>%d</code>", userID)
	}

	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if user.Username != "" {
		return fmt.Sprintf("<b>%s</b> @%s", name, user.Username)
	}
	return fmt.Sprintf("<b>%s</b> (ID: <code>%d</code>)", name, userID)
}

func activateSubscription(b *bot.Bot, userID int64, paymentRef, cardToken string) error {
	existing, err := db.Query.GetSubscriptionByPaymentRef(b.Ctx, paymentRef)
	if err == nil {
		log.Printf("subscription %s already exists (id=%d), skipping", paymentRef, existing.ID)
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	starts := time.Now()
	active, err := db.Query.GetActiveSubscription(b.Ctx, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err == nil {
		starts = active.Ends
	}
	ends := starts.AddDate(0, 1, 0)

	_, err = db.Query.CreateSubscription(b.Ctx, db.CreateSubscriptionParams{
		UserID:       userID,
		PaymentRef:   paymentRef,
		PaymentToken: cardToken,
		Starts:       starts,
		Ends:         ends,
		IsManual:     false,
	})
	return err
}

// RenewAllDueSubscriptions charges every subscription ending today with a saved card token.
func RenewAllDueSubscriptions(b *bot.Bot) {
	subs, err := db.Query.GetSubscriptionsForRenewal(b.Ctx)
	if err != nil {
		b.Error("get subs for renewal: " + err.Error())
		return
	}
	for _, sub := range subs {
		if err := RenewSubscription(b, sub); err != nil {
			log.Printf("renew sub %d: %v", sub.ID, err)
		}
	}
}

// RenewSubscription charges the saved Morning card token and extends the sub.
// Called by cron for subs whose `ends = CURRENT_DATE`.
func RenewSubscription(b *bot.Bot, sub db.GetSubscriptionsForRenewalRow) error {
	if sub.PaymentToken == "" {
		return fmt.Errorf("sub %d: no payment token", sub.ID)
	}

	name := strings.TrimSpace(sub.FirstName + " " + sub.LastName)
	newRef, err := chargeMorningToken(morningChargeRequest{
		token:       sub.PaymentToken,
		description: "Yoletta Online — monthly renewal",
		amount:      subscriptionPriceILS,
		clientName:  name,
		custom:      strconv.FormatInt(sub.UserID, 10),
	})
	if err != nil {
		if dbErr := db.Query.DeactivateSubscription(b.Ctx, sub.ID); dbErr != nil {
			b.Error("deactivate after charge fail: " + dbErr.Error())
		}
		b.SendHTML(sub.UserID, "❌ Не удалось продлить подписку автоматически. Обнови оплату через бота.")
		b.SendHTML(common.AdminChatID(), fmt.Sprintf(
			"⚠️ Auto-renew failed for user <code>%d</code>: %s",
			sub.UserID, err.Error(),
		))
		return err
	}

	newEnds := sub.Ends
	if newEnds.Before(time.Now()) {
		newEnds = time.Now()
	}
	newEnds = newEnds.AddDate(0, 1, 0)

	if err := db.Query.ExtendSubscription(b.Ctx, db.ExtendSubscriptionParams{
		ID:   sub.ID,
		Ends: newEnds,
	}); err != nil {
		return err
	}

	b.SendHTML(sub.UserID, fmt.Sprintf(
		"Подписка продлена автоматически до <b>%s</b> 🎉",
		newEnds.Format("02-01-06"),
	))
	b.SendHTML(common.AdminChatID(), fmt.Sprintf(
		"🔄 Auto-renewal: user <code>%d</code> extended to %s (payment <code>%s</code>)",
		sub.UserID, newEnds.Format("02-01-06"), newRef,
	))
	return nil
}

type morningChargeRequest struct {
	token       string
	description string
	amount      float64
	clientName  string
	custom      string
}

// chargeMorningToken hits Morning's tokenized charge endpoint using a previously saved card token.
// Returns the new payment ID.
func chargeMorningToken(req morningChargeRequest) (string, error) {
	token, err := getMorningToken()
	if err != nil {
		return "", err
	}

	body := map[string]any{
		"description": req.description,
		"type":        400,
		"lang":        "he",
		"currency":    "ILS",
		"vatType":     0,
		"amount":      req.amount,
		"maxPayments": 1,
		"cardToken":   req.token,
		"client":      map[string]any{"name": req.clientName},
		"custom":      req.custom,
	}

	data, _ := json.Marshal(body)
	httpReq, _ := http.NewRequest(http.MethodPost, morningBase()+"/payments/charge", bytes.NewBuffer(data))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	log.Printf("morning charge: %s", raw)

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("morning charge: status %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		ID     string `json:"id"`
		Status int    `json:"status"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}

	if result.Status != morningPaymentSuccess {
		return "", fmt.Errorf("morning charge: unexpected status %d", result.Status)
	}
	return result.ID, nil
}
