package service

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"os"
)

type PasswordResetMailer interface {
	SendResetLink(email string, resetURL string) error
}

type SMTPPasswordResetMailer struct{}

func NewSMTPPasswordResetMailer() PasswordResetMailer {
	return &SMTPPasswordResetMailer{}
}

func (m *SMTPPasswordResetMailer) SendResetLink(
	email string,
	resetURL string,
) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_AUTH_EMAIL")
	password := os.Getenv("SMTP_AUTH_PASSWORD")

	if host == "" || port == "" ||
		username == "" || password == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	recipient, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid recipient email: %w", err)
	}

	from := mail.Address{
		Name:    "Story2Gherkin",
		Address: username,
	}

	message := "From: " + from.String() + "\r\n" +
		"To: " + recipient.String() + "\r\n" +
		"Subject: Reset Your Story2Gherkin Password\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		"We received a request to reset your password.\r\n\r\n" +
		"Click the link below to reset your password:\r\n" +
		resetURL + "\r\n\r\n" +
		"If you did not request this, ignore this email.\r\n" +
		"This link expires in 30 minutes.\r\n"

	auth := smtp.PlainAuth(
		"",
		username,
		password,
		host,
	)

	return smtp.SendMail(
		net.JoinHostPort(host, port),
		auth,
		username,
		[]string{recipient.Address},
		[]byte(message),
	)
}
