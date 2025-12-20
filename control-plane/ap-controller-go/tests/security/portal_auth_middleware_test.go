package security_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ap-controller-go/internal/security"
)

// Store defines the minimal interface required by PortalAuthMiddleware.
type Store interface {
	SetNX(ctx context.Context, key string, val string, ttl time.Duration) (bool, error)
	RawKey(parts ...string) string
}

type fakeStore struct{}

func (f *fakeStore) SetNX(ctx context.Context, key, val string, ttl time.Duration) (bool, error) {
	return true, nil
}

func (f *fakeStore) RawKey(parts ...string) string {
	return "test-key"
}

// --------------------------------

func TestPortalAuthMiddleware_ContextInjection(t *testing.T) {

	security.SkipAuthForTest = true
	defer func() {
		security.SkipAuthForTest = false
	}()

	// 1. fake store
	st := &fakeStore{}

	// 2. middleware
	mw := security.PortalAuthMiddleware(st)

	// 3. downstream handler
	var gotMAC string
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(security.CtxKeyClientMAC)
		if v == nil {
			t.Fatalf("client MAC not found in context")
		}
		gotMAC = v.(string)
		w.WriteHeader(http.StatusOK)
	}))

	// 4. request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-MAC", "aa:bb:cc:dd:ee:ff")

	// 以下 header 只是为了让 middleware 不提前 reject
	req.Header.Set("X-Portal-Timestamp", "1234567890")
	req.Header.Set("X-Portal-Nonce", "test-nonce")
	req.Header.Set("X-Portal-Signature", "dummy")

	rr := httptest.NewRecorder()

	// 5. run
	h.ServeHTTP(rr, req)

	// 6. assert
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", rr.Code)
	}
	if gotMAC != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("unexpected mac: %s", gotMAC)
	}
}
