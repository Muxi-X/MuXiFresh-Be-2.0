package email

import (
	"crypto/tls"
	"errors"
	"net/smtp"
	"strings"
	"testing"

	jordanemail "github.com/jordan-wright/email"
)

func TestSendUsesImplicitTLS(t *testing.T) {
	previousInfo := eInfo
	previousSendWithTLS := sendWithTLS
	t.Cleanup(func() {
		eInfo = previousInfo
		sendWithTLS = previousSendWithTLS
	})

	eInfo = EmailInfo{
		Host:     "smtp.example.com",
		Port:     "465",
		UserName: "sender@example.com",
		Password: "secret",
	}

	expectedError := errors.New("stop before delivery")
	var capturedAddress string
	var capturedConfig *tls.Config
	sendWithTLS = func(message *jordanemail.Email, address string, auth smtp.Auth, config *tls.Config) error {
		if message.From != eInfo.UserName {
			t.Fatalf("unexpected sender: %q", message.From)
		}
		if len(message.To) != 1 || message.To[0] != "recipient@example.com" {
			t.Fatalf("unexpected recipients: %v", message.To)
		}
		if auth == nil {
			t.Fatal("SMTP auth must be configured")
		}
		capturedAddress = address
		capturedConfig = config
		return expectedError
	}

	err := Send("recipient@example.com", "set_password", "ABC123")
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected send error %v, got %v", expectedError, err)
	}
	if capturedAddress != "smtp.example.com:465" {
		t.Fatalf("unexpected SMTP address: %q", capturedAddress)
	}
	if capturedConfig == nil {
		t.Fatal("TLS config must be provided")
	}
	if capturedConfig.ServerName != "smtp.example.com" {
		t.Fatalf("unexpected TLS server name: %q", capturedConfig.ServerName)
	}
	if capturedConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("unexpected minimum TLS version: %d", capturedConfig.MinVersion)
	}
	if capturedConfig.InsecureSkipVerify {
		t.Fatal("TLS certificate verification must remain enabled")
	}
}

func TestSendEmbedsProvidedCode(t *testing.T) {
	previousInfo := eInfo
	previousSendWithTLS := sendWithTLS
	t.Cleanup(func() {
		eInfo = previousInfo
		sendWithTLS = previousSendWithTLS
	})

	eInfo = EmailInfo{
		Host:     "smtp.example.com",
		Port:     "465",
		UserName: "sender@example.com",
		Password: "secret",
	}

	const randCode = "CODE42"
	var capturedBody string
	sendWithTLS = func(message *jordanemail.Email, address string, auth smtp.Auth, config *tls.Config) error {
		capturedBody = string(message.HTML)
		return nil
	}

	if err := Send("recipient@example.com", "set_password", randCode); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
	if !strings.Contains(capturedBody, randCode) {
		t.Fatal("delivered email must contain the provided verification code")
	}
}

func TestSendRejectsInvalidType(t *testing.T) {
	previousInfo := eInfo
	t.Cleanup(func() { eInfo = previousInfo })

	eInfo = EmailInfo{Host: "smtp.example.com", Port: "465"}

	err := Send("recipient@example.com", "unknown_type", "CODE42")
	if !errors.Is(err, ErrInvalidEmailType) {
		t.Fatalf("expected %v, got %v", ErrInvalidEmailType, err)
	}
}

func TestSendRejectsMissingCode(t *testing.T) {
	previousInfo := eInfo
	t.Cleanup(func() { eInfo = previousInfo })

	eInfo = EmailInfo{Host: "smtp.example.com", Port: "465"}

	err := Send("recipient@example.com", "set_password", "")
	if !errors.Is(err, ErrMissingEmailCode) {
		t.Fatalf("expected %v, got %v", ErrMissingEmailCode, err)
	}
}
