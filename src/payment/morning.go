package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func morningBase() string {
	if os.Getenv("ENV") == "production" {
		return "https://api.greeninvoice.co.il/api/v1"
	}
	return "https://sandbox.d.greeninvoice.co.il/api/v1"
}

func getMorningToken() (string, error) {
	payload := map[string]string{
		"id":     os.Getenv("MORNING_API_KEY"),
		"secret": os.Getenv("MORNING_SECRET"),
	}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(morningBase()+"/account/token", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Token == "" {
		return "", fmt.Errorf("morning: empty token")
	}
	return result.Token, nil
}

func createMorningReceipt(c captureResult) error {
	plan, ok := plans[c.PlanKey]
	if !ok {
		return fmt.Errorf("morning: unknown plan %q", c.PlanKey)
	}

	token, err := getMorningToken()
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")

	client := map[string]any{
		"name": c.Name,
	}
	if c.Email != "" {
		client["emails"] = []string{c.Email}
	}

	body := map[string]any{
		"type":     400,
		"lang":     "he",
		"currency": "ILS",
		"vatType":  0,
		"remarks":  plan.Label,
		"client":   client,
		"payment": []map[string]any{{
			"type":     5,
			"price":    plan.Price,
			"date":     today,
			"bankName": "PayPal",
		}},
	}

	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, morningBase()+"/documents", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("morning: status %d: %v", resp.StatusCode, errBody)
	}
	return nil
}
