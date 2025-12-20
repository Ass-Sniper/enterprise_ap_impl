package security

import (
	"context"
	"time"
)

type Store interface {
	SetNX(ctx context.Context, key string, val string, ttl time.Duration) (bool, error)
	RawKey(parts ...string) string
}
