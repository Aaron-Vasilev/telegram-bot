package payment

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/online/db"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Provider is a payment service that sells the online club subscription.
// Prodamus is the only one for now, more (e.g. for foreign cards) can be added later.
type Provider interface {
	// Name is a stable id, used in webhook paths and stored payment refs.
	Name() db.OnlinePaymentProvider
	// Button is the label of the purchase button in the bot.
	Button() string
	PaymentLink(userId int64) (string, error)
	ParseWebhook(r *http.Request) (Event, error)
}

// Event is a provider webhook translated to what the bot cares about.
type Event struct {
	Paid bool
	// PaymentRef is unique per charge, used to skip duplicated webhooks.
	PaymentRef string
	// CustomerRef is a stable payer id at the provider, survives renewals.
	CustomerRef string
	// UserID is the Telegram ID, 0 when the provider didn't send it back.
	UserID   int64
	Amount   string
	Currency string
	Status   string
}

var (
	// ErrIgnored marks webhooks that aren't about the club subscription.
	ErrIgnored      = errors.New("webhook ignored")
	ErrBadSignature = errors.New("bad signature")
)

var providers []Provider

func Providers() []Provider {
	return providers
}

func StartPaymentServer(b *bot.Bot) {
	if p, ok := newProdamus(); ok {
		providers = append(providers, p)
	} else {
		log.Println("Prodamus is not configured")
	}

	for _, p := range providers {
		b.Mux.HandleFunc("/api/online/"+string(p.Name())+"/webhook", webhookHandler(b, p))
	}
}

func webhookHandler(b *bot.Bot, p Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		event, err := p.ParseWebhook(r)
		if errors.Is(err, ErrIgnored) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if err != nil {
			b.Error(fmt.Sprintf("%s webhook: %s", p.Name(), err.Error()))
			http.Error(w, "invalid webhook", http.StatusBadRequest)
			return
		}

		if err := handleEvent(b, p, event); err != nil {
			b.Error(fmt.Sprintf("%s webhook %s: %s", p.Name(), event.PaymentRef, err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleEvent(b *bot.Bot, p Provider, event Event) error {
	if !event.Paid {
		b.SendHTML(common.AdminChatID(), fmt.Sprintf(
			"⚠️ %s: платёж не прошёл\n%s\nСтатус: %s\nПлатёж: <code>%s</code>",
			p.Name(), formatUser(b, event.UserID), event.Status, event.PaymentRef,
		))
		return nil
	}

	userId := event.UserID
	if userId == 0 && event.CustomerRef != "" {
		id, err := db.Query.GetCustomerPayment(b.Ctx, db.GetCustomerPaymentParams{
			Provider:   p.Name(),
			ExternalID: event.CustomerRef,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		userId = id
	}

	if userId == 0 {
		b.SendHTML(common.AdminChatID(), fmt.Sprintf(
			"⚠️ %s: оплата %s %s, но непонятно от кого. Выдай подписку вручную.\nПлатёж: <code>%s</code>\nПокупатель: <code>%s</code>",
			p.Name(), event.Amount, event.Currency, event.PaymentRef, event.CustomerRef,
		))
		return nil
	}

	if event.CustomerRef != "" {
		if err := db.Query.UpsertCustomerPayment(b.Ctx, db.UpsertCustomerPaymentParams{
			Provider:   p.Name(),
			ExternalID: event.CustomerRef,
			UserID:     userId,
		}); err != nil {
			return err
		}
	}

	paymentRef := string(p.Name()) + ":" + event.PaymentRef
	existing, err := db.Query.GetSubscriptionByPaymentRef(b.Ctx, paymentRef)
	if err == nil {
		log.Printf("subscription %s already exists (id=%d), skipping", paymentRef, existing.ID)
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	starts := time.Now()
	active, err := db.Query.GetActiveSubscription(b.Ctx, userId)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err == nil {
		starts = active.Ends
	}
	ends := starts.AddDate(0, 1, 0)

	if _, err := db.Query.CreateSubscription(b.Ctx, db.CreateSubscriptionParams{
		UserID:     userId,
		PaymentRef: paymentRef,
		Starts:     starts,
		Ends:       ends,
		IsManual:   false,
	}); err != nil {
		return err
	}

	b.SendHTML(userId, fmt.Sprintf("Оплата прошла, подписка активна до <b>%s</b> 🎊", ends.Format("02-01-06")))
	b.SendHTML(common.AdminChatID(), fmt.Sprintf(
		"✅ %s: оплата %s %s\n%s\nПодписка до %s\nПлатёж: <code>%s</code>",
		p.Name(), event.Amount, event.Currency, formatUser(b, userId), ends.Format("02-01-06"), event.PaymentRef,
	))
	return nil
}

func formatUser(b *bot.Bot, userId int64) string {
	user, err := db.Query.GetUser(b.Ctx, userId)
	if err != nil {
		return fmt.Sprintf("Telegram ID: <code>%d</code>", userId)
	}

	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if user.Username != "" {
		return fmt.Sprintf("<b>%s</b> @%s", name, user.Username)
	}
	return fmt.Sprintf("<b>%s</b> (ID: <code>%d</code>)", name, userId)
}
