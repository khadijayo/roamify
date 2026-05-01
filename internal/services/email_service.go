package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const resendEmailsURL = "https://api.resend.com/emails"

func SendVerificationEmail(to string, token string) error {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_BASE_URL")), "/")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is required")
	}
	if baseURL == "" {
		return fmt.Errorf("APP_BASE_URL is required")
	}

	body := map[string]interface{}{
		"from":    "onboarding@resend.dev",
		"to":      []string{to},
		"subject": "Verify your email",
		"html": fmt.Sprintf(
			"<h2>Verify your email</h2><p>Click below:</p><a href='%s/api/v1/auth/verify?token=%s'>Verify Email</a>",
			baseURL,
			token,
		),
	}

	return sendResendEmail(apiKey, body)
}

func SendPasswordResetCode(to string, fullName string, code string) error {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is required")
	}

	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "traveler"
	}

	body := map[string]interface{}{
		"from":    "onboarding@resend.dev",
		"to":      []string{to},
		"subject": "Reset your Roamify password",
		"html": fmt.Sprintf(
			"<p>Hi %s,</p><p>Use the code below to reset your Roamify password:</p><p><strong>%s</strong></p><p>This code expires in 10 minutes and can only be used once.</p>",
			name,
			code,
		),
	}

	return sendResendEmail(apiKey, body)
}

func sendResendEmail(apiKey string, body map[string]interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, resendEmailsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("resend API error: %d %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}

	return nil
}
