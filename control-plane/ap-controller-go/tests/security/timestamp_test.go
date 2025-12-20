package security_test

import (
	"testing"
	"time"

	"ap-controller-go/internal/security"
)

func TestValidateTimestamp_OK(t *testing.T) {
	now := time.Unix(1700000000, 0)
	ts := "1700000000"

	if err := security.ValidateTimestamp(ts, now); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestValidateTimestamp_TooOld(t *testing.T) {
	now := time.Unix(1700000000, 0)
	ts := "1699990000"

	if err := security.ValidateTimestamp(ts, now); err == nil {
		t.Fatalf("expected error")
	}
}

func TestValidateTimestamp_TooNew(t *testing.T) {
	now := time.Unix(1700000000, 0)
	ts := "1700001000"

	if err := security.ValidateTimestamp(ts, now); err == nil {
		t.Fatalf("expected error")
	}
}

func TestValidateTimestamp_Invalid(t *testing.T) {
	now := time.Unix(1700000000, 0)

	if err := security.ValidateTimestamp("abc", now); err == nil {
		t.Fatalf("expected error")
	}
}
