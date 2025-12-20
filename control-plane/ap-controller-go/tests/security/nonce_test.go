package security_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"ap-controller-go/internal/security"
)

type fakeNonceStore struct {
	seen map[string]bool
}

func newFakeNonceStore() *fakeNonceStore {
	return &fakeNonceStore{seen: make(map[string]bool)}
}

func (f *fakeNonceStore) SetNX(ctx context.Context, key, val string, ttl time.Duration) (bool, error) {
	if f.seen[key] {
		return false, nil
	}
	f.seen[key] = true
	return true, nil
}

func (f *fakeNonceStore) RawKey(parts ...string) string {
	return "test:" + strings.Join(parts, ":")
}

func TestValidateNonce_OK(t *testing.T) {
	st := newFakeNonceStore()

	err := security.ValidateNonce(context.Background(), st, "nonce-1")
	if err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestValidateNonce_Replay(t *testing.T) {
	st := newFakeNonceStore()

	_ = security.ValidateNonce(context.Background(), st, "nonce-1")

	if err := security.ValidateNonce(context.Background(), st, "nonce-1"); err == nil {
		t.Fatalf("expected replay error")
	}
}

func TestValidateNonce_Empty(t *testing.T) {
	st := newFakeNonceStore()

	if err := security.ValidateNonce(context.Background(), st, ""); err == nil {
		t.Fatalf("expected error")
	}
}
