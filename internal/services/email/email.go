package email

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/khadijayo/roamify/config"
)

type Message struct {
	To       []string
	Subject  string
	TextBody string
}

type Service interface {
	Send(ctx context.Context, msg Message) error
	SendVerificationEmail(ctx context.Context, to, fullName, verificationURL string) error
}

type service struct {
	host      string
	port      string
	username  string
	password  string
	fromEmail string
	fromName  string
}

func NewService(cfg *config.Config) Service {
	return &service{
		host:      cfg.SMTPHost,
		port:      cfg.SMTPPort,
		username:  cfg.SMTPUsername,
		password:  cfg.SMTPPassword,
		fromEmail: cfg.SMTPFromEmail,
		fromName:  cfg.SMTPFromName,
	}
}

func (s *service) Send(ctx context.Context, msg Message) error {
	if len(msg.To) == 0 {
		return errors.New("email recipient is required")
	}
	if s.host == "" || s.port == "" || s.fromEmail == "" {
		return errors.New("smtp configuration is incomplete")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	payload := buildMessage(s.fromEmail, s.fromName, msg)

	return smtp.SendMail(addr, auth, s.fromEmail, msg.To, []byte(payload))
}

func (s *service) SendVerificationEmail(ctx context.Context, to, fullName, verificationURL string) error {
	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "traveler"
	}

	body := fmt.Sprintf(
		"Hi %s,\n\nVerify your Roamify account by opening the link below:\n%s\n\nThis link expires in 15 minutes. If you did not create this account, you can ignore this email.\n",
		name,
		verificationURL,
	)

	return s.Send(ctx, Message{
		To:       []string{to},
		Subject:  "Verify your Roamify account",
		TextBody: body,
	})
}

func buildMessage(fromEmail, fromName string, msg Message) string {
	displayFrom := fromEmail
	if strings.TrimSpace(fromName) != "" {
		displayFrom = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	}

	return strings.Join([]string{
		fmt.Sprintf("From: %s", displayFrom),
		fmt.Sprintf("To: %s", strings.Join(msg.To, ", ")),
		fmt.Sprintf("Subject: %s", msg.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		msg.TextBody,
	}, "\r\n")
}
