package security

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

func CanonicalString(req *http.Request, body []byte) string {
	h := sha256.Sum256(body)
	bodyHash := hex.EncodeToString(h[:])

	return req.Method + "\n" +
		req.URL.Path + "\n" +
		req.URL.RawQuery + "\n" +
		bodyHash + "\n"
}
