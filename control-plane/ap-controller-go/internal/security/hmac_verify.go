package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
)

var (
	ErrNotInitialized = errors.New("portal hmac not initialized")
	ErrInvalidSign    = errors.New("invalid hmac signature")
)

func VerifyPortalSignature(req *http.Request, body []byte) error {
	ks := PortalHMACProvider()
	if ks == nil {
		return ErrNotInitialized
	}

	// --------------------------------------------------
	// Headers
	// --------------------------------------------------
	kid := req.Header.Get("X-Portal-Kid") // optional
	ts := req.Header.Get("X-Portal-Timestamp")
	nonce := req.Header.Get("X-Portal-Nonce")
	sign := req.Header.Get("X-Portal-Signature")

	if ts == "" || nonce == "" || sign == "" {
		return ErrInvalidSign
	}

	// --------------------------------------------------
	// Select key
	// --------------------------------------------------
	if kid == "" {
		kid = ks.CurrentKID
	}

	key, ok := ks.Keys[kid]
	if !ok || key == nil {
		return ErrInvalidSign
	}

	// --------------------------------------------------
	// Build canonical string
	// --------------------------------------------------
	canonical := ts + "\n" + nonce + "\n" + CanonicalString(req, body)

	// --------------------------------------------------
	// HMAC-SHA256 verification
	// --------------------------------------------------
	expectMAC := hmac.New(sha256.New, key)
	expectMAC.Write([]byte(canonical))
	expected := expectMAC.Sum(nil)

	actual, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return ErrInvalidSign
	}

	if !hmac.Equal(expected, actual) {
		return ErrInvalidSign
	}

	return nil
}
