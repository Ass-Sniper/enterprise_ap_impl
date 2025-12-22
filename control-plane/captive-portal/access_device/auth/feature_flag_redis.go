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
	// 全局开关, Feature = 业务域 + 能力
	globalGrainKey := "feature:auth:" + feature
	if ok, _ := f.RDB.Get(ctx, globalGrainKey).Bool(); ok {
		return true
	}

	// 用户灰度
	userGrainKey := fmt.Sprintf("feature:auth:%s:user:%s", feature, user)
	ok, _ := f.RDB.Get(ctx, userGrainKey).Bool()
	return ok
}
