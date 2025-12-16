package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ap-controller-go/internal/config"

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

func (s *Store) SetSession(ctx context.Context, sess SessionV2, ttlSec int) error {
	now := time.Now().Unix()
	if sess.Schema == 0 {
		sess.Schema = 2
	}
	if sess.TS.Created == 0 {
		sess.TS.Created = now
	}
	sess.TS.Updated = now

	b, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(sess.MAC), string(b), time.Duration(ttlSec)*time.Second).Err()
}

func (s *Store) GetSessionFull(ctx context.Context, mac string) (*SessionV2, int, error) {
	k := s.key(mac)
	val, err := s.rdb.Get(ctx, k).Result()
	if err == redis.Nil {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	var sess SessionV2
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return nil, 0, err
	}
	ttl, err := s.rdb.TTL(ctx, k).Result()
	if err != nil {
		return &sess, 0, nil
	}
	ttlSec := int(ttl / time.Second)
	if ttlSec < 0 {
		ttlSec = 0
	}
	return &sess, ttlSec, nil
}

func (s *Store) Refresh(ctx context.Context, mac string, ttlSec int) (bool, error) {
	k := s.key(mac)
	ok, err := s.rdb.Expire(ctx, k, time.Duration(ttlSec)*time.Second).Result()
	if err == redis.Nil {
		return false, nil
	}
	return ok, err
}

func (s *Store) Delete(ctx context.Context, mac string) (bool, error) {
	n, err := s.rdb.Del(ctx, s.key(mac)).Result()
	return n > 0, err
}
