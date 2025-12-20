package store

import (
	"ap-controller-go/internal/config"
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

func New(cfg *config.Config, password string) *Store {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       cfg.Redis.DB,
	})
	return &Store{
		cfg:    cfg,
		rdb:    rdb,
		prefix: cfg.Redis.Prefix,
	}
}

func (s *Store) key(mac string) string { return s.prefix + mac }

func (s *Store) Ping(ctx context.Context) error {
	return s.rdb.Ping(ctx).Err()
}

func (s *Store) RawKey(parts ...string) string {
	return s.prefix + strings.Join(parts, ":")
}
