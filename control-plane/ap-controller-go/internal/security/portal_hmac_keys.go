package security

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	secretDir     = "/run/secrets"
	secretPrefix  = "portal_hmac_"
	envCurrentKID = "PORTAL_HMAC_CURRENT_KID"
	envFallback   = "PORTAL_HMAC_SECRET"
)

type KeySet struct {
	CurrentKID string
	Keys       map[string][]byte
}

var portalHMAC *KeySet

func InitPortalHMAC(ks *KeySet) {
	portalHMAC = ks
}

func PortalHMAC() *KeySet {
	return portalHMAC
}

func LoadPortalHMACKeySet() (*KeySet, error) {
	ks := &KeySet{
		Keys: make(map[string][]byte),
	}

	// current kid
	ks.CurrentKID = os.Getenv(envCurrentKID)
	if ks.CurrentKID == "" {
		ks.CurrentKID = "v1"
	}

	// load from /run/secrets
	if entries, err := os.ReadDir(secretDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !strings.HasPrefix(e.Name(), secretPrefix) {
				continue
			}

			kid := strings.TrimPrefix(e.Name(), secretPrefix)
			path := filepath.Join(secretDir, e.Name())

			key, err := readBase64File(path)
			if err != nil {
				return nil, fmt.Errorf("load %s failed: %w", path, err)
			}
			ks.Keys[kid] = key
		}
	}

	// fallback: env (dev only)
	if len(ks.Keys) == 0 {
		if env := os.Getenv(envFallback); env != "" {
			key, err := base64.StdEncoding.DecodeString(env)
			if err != nil {
				return nil, err
			}
			ks.Keys[ks.CurrentKID] = key
		}
	}

	if len(ks.Keys) == 0 {
		return nil, errors.New("no portal hmac secret found")
	}
	if _, ok := ks.Keys[ks.CurrentKID]; !ok {
		return nil, fmt.Errorf("current kid %s not found", ks.CurrentKID)
	}

	return ks, nil
}

func readBase64File(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(strings.TrimSpace(string(b)))
}
