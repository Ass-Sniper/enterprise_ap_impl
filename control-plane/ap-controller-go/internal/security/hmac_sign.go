package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Signature struct {
	KID       string
	Timestamp string
	Nonce     string
	Signature string
}

func SignPortalRequest(req *http.Request, body []byte) (*Signature, error) {
	ks := PortalHMACProvider()
	if ks == nil {
		return nil, ErrNotInitialized
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := uuid.NewString()

	canonical := ts + "\n" + nonce + "\n" + CanonicalString(req, body)

	key := ks.Keys[ks.CurrentKID]
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(canonical))

	return &Signature{
		KID:       ks.CurrentKID,
		Timestamp: ts,
		Nonce:     nonce,
		Signature: base64.StdEncoding.EncodeToString(mac.Sum(nil)),
	}, nil
}
