package security_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"testing"

	"ap-controller-go/internal/security"
)

func signRequest(
	req *http.Request,
	body []byte,
	secret []byte,
) string {
	ts := req.Header.Get("X-Portal-Timestamp")
	nonce := req.Header.Get("X-Portal-Nonce")

	canonical := ts + "\n" + nonce + "\n" + security.CanonicalString(req, body)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(canonical))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestVerifyPortalSignature_OK(t *testing.T) {
	// --- inject test KeySet ---
	orig := security.PortalHMACProvider
	security.PortalHMACProvider = func() *security.KeySet {
		return &security.KeySet{
			CurrentKID: "k1",
			Keys: map[string][]byte{
				"k1": []byte("test-secret"),
			},
		}
	}
	defer func() { security.PortalHMACProvider = orig }()

	body := []byte(`{"hello":"world"}`)
	req, _ := http.NewRequest("POST", "/portal/context/verify", bytes.NewReader(body))

	req.Header.Set("X-Portal-Timestamp", "1700000000")
	req.Header.Set("X-Portal-Nonce", "nonce-1")

	sign := signRequest(req, body, []byte("test-secret"))
	req.Header.Set("X-Portal-Signature", sign)

	if err := security.VerifyPortalSignature(req, body); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestVerifyPortalSignature_BadSignature(t *testing.T) {
	orig := security.PortalHMACProvider
	security.PortalHMACProvider = func() *security.KeySet {
		return &security.KeySet{
			CurrentKID: "k1",
			Keys: map[string][]byte{
				"k1": []byte("test-secret"),
			},
		}
	}
	defer func() { security.PortalHMACProvider = orig }()

	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("X-Portal-Timestamp", "1700000000")
	req.Header.Set("X-Portal-Nonce", "nonce-1")
	req.Header.Set("X-Portal-Signature", "invalid==")

	if err := security.VerifyPortalSignature(req, nil); err == nil {
		t.Fatalf("expected error")
	}
}
