package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const resendAPIURL = "https://api.resend.com/emails"

type resendRequest struct {
	From    string `json:"from"`
	To      []string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

type resendResponse struct {
	ID    string `json:"id"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type resendErrorResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func SendVerificationEmailResend(to string, token string) error {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_BASE_URL")), "/")
	if baseURL == "" {
		return fmt.Errorf("APP_BASE_URL is required")
	}

	verifyURL := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", baseURL, token)

	htmlBody := fmt.Sprintf(
		"<h2>Verify your email</h2><p>Click below to verify your Roamify account:</p><p><a href=\"%s\" style=\"background-color:#007bff;color:white;padding:10px 20px;text-decoration:none;border-radius:5px;display:inline-block;\">Verify Email</a></p><p>This link expires in 24 hours.</p>",
		verifyURL,
	)

	textBody := fmt.Sprintf(
		"Verify your Roamify account by opening this link: %s\n\nThis link expires in 24 hours.",
		verifyURL,
	)

	fromEmail := strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL"))
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
	}

	return sendResendEmail(resendRequest{
		From:    fromEmail,
		To:      []string{to},
		Subject: "Verify your email - Roamify",
		HTML:    htmlBody,
		Text:    textBody,
	})
}

func SendPasswordResetCodeResend(to string, fullName string, code string) error {
	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "traveler"
	}

	htmlBody := fmt.Sprintf(
		"<p>Hi %s,</p><p>Use the code below to reset your Roamify password:</p><p><strong style=\"font-size:24px;letter-spacing:2px;\">%s</strong></p><p>This code expires in 10 minutes and can only be used once.</p>",
		name,
		code,
	)

	textBody := fmt.Sprintf(
		"Hi %s,\n\nUse this code to reset your Roamify password: %s\n\nThis code expires in 10 minutes and can only be used once.",
		name,
		code,
	)

	fromEmail := strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL"))
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
	}

	return sendResendEmail(resendRequest{
		From:    fromEmail,
		To:      []string{to},
		Subject: "Reset your Roamify password",
		HTML:    htmlBody,
		Text:    textBody,
	})
}

func sendResendEmail(req resendRequest) error {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is required")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal resend request: %w", err)
	}

	log.Printf("[email] sending via Resend to=%s subject=%q", strings.Join(req.To, ","), req.Subject)

	httpReq, err := http.NewRequest("POST", resendAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("resend request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read resend response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var resendErr resendErrorResponse
		if err := json.Unmarshal(body, &resendErr); err == nil {
			if strings.TrimSpace(resendErr.Message) != "" {
				return fmt.Errorf("resend api error (status %d): %s", resp.StatusCode, resendErr.Message)
			}
			if strings.TrimSpace(resendErr.Name) != "" {
				return fmt.Errorf("resend api error (status %d): %s", resp.StatusCode, resendErr.Name)
			}
		}
		bodyText := strings.TrimSpace(string(body))
		if bodyText == "" {
			bodyText = resp.Status
		}
		return fmt.Errorf("resend api error (status %d): %s (content-type=%s)", resp.StatusCode, bodyText, resp.Header.Get("Content-Type"))
	}

	var resendResp resendResponse
	if err := json.Unmarshal(body, &resendResp); err != nil {
		return fmt.Errorf("failed to parse resend response: %w", err)
	}

	log.Printf("[email] resend email sent successfully to=%s id=%s", strings.Join(req.To, ","), resendResp.ID)
	return nil
}
