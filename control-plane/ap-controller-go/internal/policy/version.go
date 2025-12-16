package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// BuildPolicyVersion calculates version metadata for runtime policy
func BuildPolicyVersion(payload any, baseVersion string) PolicyVersion {
	raw, _ := json.Marshal(payload)

	sum := sha256.Sum256(raw)

	return PolicyVersion{
		Version:   baseVersion,
		Checksum:  hex.EncodeToString(sum[:]),
		Generated: time.Now().Unix(),
	}
}
