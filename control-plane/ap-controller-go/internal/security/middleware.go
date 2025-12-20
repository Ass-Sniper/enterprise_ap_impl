package security

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

const (
	maxSkewSeconds = 300
)

// SkipAuthForTest disables auth checks in tests.
// DO NOT enable this in production.
var SkipAuthForTest = false

func PortalAuthMiddleware(st Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// --------------------------------------------------
			// Test shortcut
			// --------------------------------------------------
			if SkipAuthForTest {
				mac := r.Header.Get("X-Client-MAC")
				if mac == "" {
					http.Error(w, "missing client mac", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), CtxKeyClientMAC, mac)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			body := readBodyAndRestore(r)

			// 1. HMAC verify
			if err := VerifyPortalSignature(r, body); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// 2. timestamp window
			if err := ValidateTimestamp(
				r.Header.Get("X-Portal-Timestamp"),
				time.Now(),
			); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// 3. nonce replay protection
			if err := ValidateNonce(
				context.Background(),
				st,
				r.Header.Get("X-Portal-Nonce"),
			); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// --------------------------------------------------
			// 4. Extract client MAC and inject into context
			// --------------------------------------------------
			mac := r.Header.Get("X-Client-MAC")
			if mac == "" {
				http.Error(w, "missing client mac", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxKeyClientMAC, mac)

			// --------------------------------------------------
			// 5. Continue with enriched context
			// --------------------------------------------------
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// readBodyAndRestore reads body for HMAC and restores it for handlers
func readBodyAndRestore(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	return b
}
