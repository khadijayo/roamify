package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"
)

type emailMessage struct {
	To       []string
	Subject  string
	TextBody string
	HTMLBody string
}

func SendVerificationEmail(to string, token string) error {
	// Prefer Resend if API key is configured
	if apiKey := os.Getenv("RESEND_API_KEY"); strings.TrimSpace(apiKey) != "" {
		return SendVerificationEmailResend(to, token)
	}

	// Fall back to SMTP
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_BASE_URL")), "/")
	if baseURL == "" {
		return errors.New("APP_BASE_URL is required")
	}

	verifyURL := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", baseURL, url.QueryEscape(token))
	return sendSMTP(emailMessage{
		To:      []string{to},
		Subject: "Verify your email",
		TextBody: fmt.Sprintf(
			"Verify your Roamify account by opening this link: %s\n\nThis link expires in 24 hours.",
			verifyURL,
		),
		HTMLBody: fmt.Sprintf(
			"<h2>Verify your email</h2><p>Click below:</p><p><a href=\"%s\">Verify Email</a></p><p>This link expires in 24 hours.</p>",
			verifyURL,
		),
	})
}

func SendPasswordResetCode(to string, fullName string, code string) error {
	// Prefer Resend if API key is configured
	if apiKey := os.Getenv("RESEND_API_KEY"); strings.TrimSpace(apiKey) != "" {
		return SendPasswordResetCodeResend(to, fullName, code)
	}

	// Fall back to SMTP
	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "traveler"
	}

	return sendSMTP(emailMessage{
		To:      []string{to},
		Subject: "Reset your Roamify password",
		TextBody: fmt.Sprintf(
			"Hi %s,\n\nUse this code to reset your Roamify password: %s\n\nThis code expires in 10 minutes and can only be used once.",
			name,
			code,
		),
		HTMLBody: fmt.Sprintf(
			"<p>Hi %s,</p><p>Use the code below to reset your Roamify password:</p><p><strong>%s</strong></p><p>This code expires in 10 minutes and can only be used once.</p>",
			name,
			code,
		),
	})
}

func sendSMTP(msg emailMessage) error {
	if len(msg.To) == 0 {
		return errors.New("email recipient is required")
	}

	cfg := smtpConfigFromEnv()
	if err := cfg.validate(); err != nil {
		return err
	}

	addr := net.JoinHostPort(cfg.host, cfg.port)
	log.Printf("[email] sending smtp message host=%s port=%s from=%s to=%s subject=%q", cfg.host, cfg.port, cfg.fromEmail, strings.Join(msg.To, ","), msg.Subject)
	conn, err := dialSMTP(addr, cfg.host, cfg.port)
	if err != nil {
		return fmt.Errorf("smtp dial failed host=%s port=%s: %w", cfg.host, cfg.port, err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))

	client, err := smtp.NewClient(conn, cfg.host)
	if err != nil {
		return fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Close()

	if cfg.port != "465" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{ServerName: cfg.host, MinVersion: tls.VersionTLS12}
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("smtp starttls failed: %w", err)
			}
		}
	}

	if cfg.username != "" && cfg.password != "" {
		auth := smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	if err := client.Mail(cfg.fromEmail); err != nil {
		return fmt.Errorf("smtp sender rejected: %w", err)
	}
	for _, recipient := range msg.To {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("smtp recipient rejected %s: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data command failed: %w", err)
	}
	if _, err := writer.Write([]byte(buildMIMEMessage(cfg, msg))); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp message write failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp message close failed: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit failed: %w", err)
	}

	log.Printf("[email] smtp message sent to=%s subject=%q", strings.Join(msg.To, ","), msg.Subject)
	return nil
}

type smtpConfig struct {
	host      string
	port      string
	username  string
	password  string
	fromEmail string
	fromName  string
}

func smtpConfigFromEnv() smtpConfig {
	fromEmail := strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL"))
	if fromEmail == "" {
		fromEmail = strings.TrimSpace(os.Getenv("SMTP_FROM"))
	}
	if fromEmail == "" {
		fromEmail = strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	}

	return smtpConfig{
		host:      strings.TrimSpace(os.Getenv("SMTP_HOST")),
		port:      envOrDefault("SMTP_PORT", "587"),
		username:  strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		password:  strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		fromEmail: fromEmail,
		fromName:  envOrDefault("SMTP_FROM_NAME", "Roamify"),
	}
}

func (c smtpConfig) validate() error {
	if c.host == "" {
		return errors.New("SMTP_HOST is required")
	}
	if c.port == "" {
		return errors.New("SMTP_PORT is required")
	}
	if c.fromEmail == "" {
		return errors.New("SMTP_FROM_EMAIL is required")
	}
	if (c.username == "") != (c.password == "") {
		return errors.New("SMTP_USERNAME and SMTP_PASSWORD must both be set")
	}
	return nil
}

func dialSMTP(addr, host, port string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	if port == "465" {
		return tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		})
	}
	return dialer.Dial("tcp", addr)
}

func buildMIMEMessage(cfg smtpConfig, msg emailMessage) string {
	body := msg.TextBody
	contentType := "text/plain; charset=UTF-8"
	if strings.TrimSpace(msg.HTMLBody) != "" {
		body = msg.HTMLBody
		contentType = "text/html; charset=UTF-8"
	}

	return strings.Join([]string{
		fmt.Sprintf("From: %s", formatFrom(cfg.fromEmail, cfg.fromName)),
		fmt.Sprintf("To: %s", strings.Join(msg.To, ", ")),
		fmt.Sprintf("Subject: %s", msg.Subject),
		"MIME-Version: 1.0",
		"Content-Type: " + contentType,
		"",
		body,
	}, "\r\n")
}

func formatFrom(email, name string) string {
	if strings.TrimSpace(name) == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
