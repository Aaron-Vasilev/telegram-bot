package payment

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/online/db"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func StartSubscriptionServer(b *bot.Bot) {
	b.Mux.HandleFunc("/api/online-config", onlineConfigHandler())
	b.Mux.HandleFunc("/api/subscription-success", subscriptionSuccessHandler(b))
	b.Mux.HandleFunc("/api/paypal-subscription", paypalSubscriptionWebhookHandler(b))

	b.Mux.Handle("/online/", http.StripPrefix("/online", http.FileServer(http.Dir("pages/online"))))
}

func onlineConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"clientId": os.Getenv("PAYPAL_CLIENT_ID"),
			"planId":   os.Getenv("PAYPAL_ONLINE_PLAN_ID"),
		})
	}
}

func subscriptionSuccessHandler(b *bot.Bot) http.HandlerFunc {
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

		subscriptionID := r.FormValue("subscription_id")
		telegramUserID := r.FormValue("telegram_user_id")

		if subscriptionID == "" || telegramUserID == "" {
			http.Error(w, "missing subscription_id or telegram_user_id", http.StatusBadRequest)
			return
		}

		tgUserId, err := strconv.ParseInt(telegramUserID, 10, 64)
		if err != nil {
			http.Error(w, "invalid telegram_user_id", http.StatusBadRequest)
			return
		}

		result, err := fetchPaypalSubscription(subscriptionID)
		if err != nil {
			b.Error("paypal fetch subscription: " + err.Error())
			http.Error(w, "verification failed", http.StatusInternalServerError)
			return
		}

		if result.Status != "ACTIVE" && result.Status != "APPROVED" {
			b.Error(fmt.Sprintf("subscription-success: unexpected status %s for %s", result.Status, subscriptionID))
			http.Error(w, "subscription not active", http.StatusBadRequest)
			return
		}

		if err := activateSubscription(b, tgUserId, subscriptionID); err != nil {
			b.Error("activateSubscription: " + err.Error())
			http.Error(w, "activation failed", http.StatusInternalServerError)
			return
		}

		b.SendHTML(common.AdminChatID(), fmt.Sprintf(
			"✅ Online subscription activated!\n%s\nPayPal sub: <code>%s</code>",
			formatUser(b, tgUserId), subscriptionID,
		))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}
}

func paypalSubscriptionWebhookHandler(b *bot.Bot) http.HandlerFunc {
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
		case "PAYMENT.SALE.COMPLETED":
			if err := handleRecurringPayment(b, event.Resource); err != nil {
				b.Error("recurring payment: " + err.Error())
			}
		case "BILLING.SUBSCRIPTION.CANCELLED", "BILLING.SUBSCRIPTION.EXPIRED", "BILLING.SUBSCRIPTION.SUSPENDED":
			subId := stringField(event.Resource, "id")
			b.SendHTML(common.AdminChatID(), fmt.Sprintf(
				"⚠️ PayPal sub %s: <code>%s</code>",
				event.EventType, subId,
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

type subscriptionResult struct {
	ID       string
	Status   string
	CustomID string
}

func fetchPaypalSubscription(subscriptionID string) (subscriptionResult, error) {
	token, err := getPaypalToken()
	if err != nil {
		return subscriptionResult{}, err
	}

	req, _ := http.NewRequest(http.MethodGet, paypalBase()+"/v1/billing/subscriptions/"+subscriptionID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return subscriptionResult{}, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	log.Printf("paypal subscription fetch: %s", raw)

	var body map[string]any
	json.Unmarshal(raw, &body)

	return subscriptionResult{
		ID:       stringField(body, "id"),
		Status:   stringField(body, "status"),
		CustomID: stringField(body, "custom_id"),
	}, nil
}

func activateSubscription(b *bot.Bot, userID int64, paypalSubID string) error {
	existing, err := db.Query.GetSubscriptionByPaypalID(b.Ctx, paypalSubID)
	if err == nil {
		log.Printf("subscription %s already exists (id=%d), skipping", paypalSubID, existing.ID)
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
		UserID:               userID,
		PaypalSubscriptionID: paypalSubID,
		Starts:               starts,
		Ends:                 ends,
		IsManual:             false,
	})
	return err
}

func handleRecurringPayment(b *bot.Bot, resource map[string]any) error {
	paypalSubID := stringField(resource, "billing_agreement_id")
	if paypalSubID == "" {
		return fmt.Errorf("missing billing_agreement_id")
	}

	sub, err := db.Query.GetSubscriptionByPaypalID(b.Ctx, paypalSubID)
	if err != nil {
		return fmt.Errorf("lookup sub %s: %w", paypalSubID, err)
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
		"🔄 Auto-renewal: user <code>%d</code> extended to %s",
		sub.UserID, newEnds.Format("02-01-06"),
	))
	return nil
}
