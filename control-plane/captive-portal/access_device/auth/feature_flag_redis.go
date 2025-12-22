package auth

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisFeatureFlag struct {
	RDB *redis.Client
}

func (f *RedisFeatureFlag) Enabled(ctx context.Context, feature, user string) bool {
	// 全局开关
	if ok, _ := f.RDB.Get(ctx, "feature:"+feature).Bool(); ok {
		return true
	}

	// 用户灰度
	key := fmt.Sprintf("feature:%s:user:%s", feature, user)
	ok, _ := f.RDB.Get(ctx, key).Bool()
	return ok
}
