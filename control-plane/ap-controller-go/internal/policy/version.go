package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// BuildControllerVersion calculates version metadata for runtime policy
func BuildControllerVersion(payload any, baseVersion string) ControllerVersion {
	raw, _ := json.Marshal(payload)

	sum := sha256.Sum256(raw)

	return ControllerVersion{
		Version:   baseVersion,
		Checksum:  hex.EncodeToString(sum[:]),
		Generated: time.Now().Unix(),
	}
}
