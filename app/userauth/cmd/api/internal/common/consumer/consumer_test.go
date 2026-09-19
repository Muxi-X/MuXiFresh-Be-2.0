package consumer

import (
	"errors"
	"testing"
	"time"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/email"
)

func withStubbedSend(t *testing.T, fn func(mailbox, typ, code string) error) {
	t.Helper()
	previousSend := sendEmail
	previousSleep := sleep
	sendEmail = fn
	sleep = func(time.Duration) {}
	t.Cleanup(func() {
		sendEmail = previousSend
		sleep = previousSleep
	})
}

func TestHandleDropsMalformedMessage(t *testing.T) {
	withStubbedSend(t, func(mailbox, typ, code string) error {
		t.Fatal("malformed message must not be sent")
		return nil
	})

	handle("not-json")
}

func TestHandleRetriesTransientFailureThenSucceeds(t *testing.T) {
	attempts := 0
	withStubbedSend(t, func(mailbox, typ, code string) error {
		attempts++
		if attempts < maxSendAttempts {
			return errors.New("temporary smtp failure")
		}
		return nil
	})

	handle(`{"email":"a@b.com","type":"set_password","rand_code":"ABC123"}`)

	if attempts != maxSendAttempts {
		t.Fatalf("expected %d attempts, got %d", maxSendAttempts, attempts)
	}
}

func TestHandleStopsAfterExhaustingRetries(t *testing.T) {
	attempts := 0
	withStubbedSend(t, func(mailbox, typ, code string) error {
		attempts++
		return errors.New("smtp down")
	})

	handle(`{"email":"a@b.com","type":"set_password","rand_code":"ABC123"}`)

	if attempts != maxSendAttempts {
		t.Fatalf("expected %d attempts, got %d", maxSendAttempts, attempts)
	}
}

func TestHandleDoesNotRetryPermanentFailure(t *testing.T) {
	attempts := 0
	withStubbedSend(t, func(mailbox, typ, code string) error {
		attempts++
		return email.ErrInvalidEmailType
	})

	handle(`{"email":"a@b.com","type":"bogus","rand_code":"ABC123"}`)

	if attempts != 1 {
		t.Fatalf("permanent failure must not be retried, got %d attempts", attempts)
	}
}

func TestHandlePassesCodeThrough(t *testing.T) {
	var gotCode string
	withStubbedSend(t, func(mailbox, typ, code string) error {
		gotCode = code
		return nil
	})

	handle(`{"email":"a@b.com","type":"set_password","rand_code":"XYZ789"}`)

	if gotCode != "XYZ789" {
		t.Fatalf("expected code XYZ789, got %q", gotCode)
	}
}
