package auth

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// RedisPolicyOverride 从 Redis 覆盖用户策略
type RedisPolicyOverride struct {
	RDB *redis.Client
}

// NewRedisPolicyOverride 创建 Redis Policy Override
func NewRedisPolicyOverride(rdb *redis.Client) *RedisPolicyOverride {
	if rdb == nil {
		return nil
	}
	return &RedisPolicyOverride{RDB: rdb}
}

// Get 获取用户策略覆盖
// key 示例：policy:user:<username>
func (r *RedisPolicyOverride) Get(
	ctx context.Context,
	username string,
) (*Policy, error) {

	key := "policy:user:" + username

	// 尝试用户策略覆盖
	val, err := r.RDB.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var p Policy
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		return nil, err
	}

	return &p, nil
}
