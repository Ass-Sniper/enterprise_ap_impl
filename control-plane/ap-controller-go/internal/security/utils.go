package security

import (
	"context"
	"errors"
	"strconv"
	"time"
)

var (
	ErrInvalidTimestamp    = errors.New("invalid timestamp")
	ErrTimestampOutOfRange = errors.New("timestamp out of range")
)

// ValidateTimestamp checks whether the given unix timestamp
// is within acceptable skew window.
func ValidateTimestamp(tsStr string, now time.Time) error {
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return ErrInvalidTimestamp
	}

	nowSec := now.Unix()
	if ts < nowSec-maxSkewSeconds || ts > nowSec+60 {
		return ErrTimestampOutOfRange
	}

	return nil
}

var ErrReplayDetected = errors.New("replay detected")

// ValidateNonce checks and records nonce to prevent replay attacks.
func ValidateNonce(ctx context.Context, st Store, nonce string) error {
	if nonce == "" {
		return ErrReplayDetected
	}

	ok, err := st.SetNX(
		ctx,
		st.RawKey("portal", "nonce", nonce),
		"1",
		10*time.Minute,
	)
	if err != nil || !ok {
		return ErrReplayDetected
	}

	return nil
}

var PortalHMACProvider = PortalHMAC
