package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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