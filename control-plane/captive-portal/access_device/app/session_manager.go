package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================
// SessionManager
// ============================

// SessionManager 负责 Portal 会话的生命周期管理
// - 创建 / 保存
// - 查询
// - 删除
// - TTL 控制
type SessionManager struct {
	rdb *redis.Client
}

// NewSessionManager 创建 SessionManager
func NewSessionManager(rdb *redis.Client) *SessionManager {
	return &SessionManager{
		rdb: rdb,
	}
}

// ============================
// Key 规范（对齐 NAC / 华为）
// ============================
//
// portal:session:{username}          → 会话主体
// portal:session:ip:{ip}             → IP → username 映射（可选）
//
// 后续可扩展：
// portal:session:mac:{mac}
// portal:session:state:{username}
//

// sessionKey 返回用户 Session Key
func sessionKey(username string) string {
	return fmt.Sprintf("portal:session:%s", username)
}

// sessionIPKey 返回 IP → username 映射 Key
func sessionIPKey(ip string) string {
	return fmt.Sprintf("portal:session:ip:%s", ip)
}

// ============================
// Session API
// ============================

// Save 保存 Session 到 Redis
func (sm *SessionManager) Save(ctx context.Context, s *Session) error {
	if sm.rdb == nil {
		return fmt.Errorf("redis client is nil")
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	ttl := time.Duration(s.TTL) * time.Second
	if ttl <= 0 {
		ttl = 30 * time.Minute // 默认兜底
	}

	pipe := sm.rdb.TxPipeline()

	pipe.Set(ctx, sessionKey(s.Username), data, ttl)
	pipe.Set(ctx, sessionIPKey(s.IP), s.Username, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	log.Printf(
		"[SESSION] save user=%s ip=%s ttl=%s policy=%s strategy=%s\n",
		s.Username, s.IP, ttl, s.Policy, s.Strategy,
	)

	return nil
}

// GetByUsername 根据用户名获取 Session
func (sm *SessionManager) GetByUsername(ctx context.Context, username string) (*Session, error) {
	val, err := sm.rdb.Get(ctx, sessionKey(username)).Result()
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// GetByIP 根据 IP 获取 Session
func (sm *SessionManager) GetByIP(ctx context.Context, ip string) (*Session, error) {
	username, err := sm.rdb.Get(ctx, sessionIPKey(ip)).Result()
	if err != nil {
		return nil, err
	}

	return sm.GetByUsername(ctx, username)
}

// Exists 判断用户是否已有有效 Session
func (sm *SessionManager) Exists(ctx context.Context, username string) (bool, error) {
	n, err := sm.rdb.Exists(ctx, sessionKey(username)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Delete 删除 Session（注销 / 过期）
func (sm *SessionManager) Delete(ctx context.Context, username string) error {
	sess, err := sm.GetByUsername(ctx, username)
	if err != nil {
		return err
	}

	pipe := sm.rdb.TxPipeline()
	pipe.Del(ctx, sessionKey(username))
	pipe.Del(ctx, sessionIPKey(sess.IP))

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	log.Printf("[SESSION] delete user=%s\n", username)
	return nil
}
