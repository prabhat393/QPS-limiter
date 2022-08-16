package mw

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// NewRedisLimiter creates new redis limiter
// Note: if expire is set to 0 it means the key has no expiration time
func NewRedisLimiter(cmdable redis.Cmdable, entity string, limit int, expire time.Duration) *RedisLimiter {

	return &RedisLimiter{
		cmdable: cmdable,
		expire:  expire,
		limit:   limit,
		key:     entity,
	}
}

// RedisLimiter struct is used for storing limit, expiration and redis client
type RedisLimiter struct {
	limit   int
	key     string
	expire  time.Duration
	cmdable redis.Cmdable
}

// Allow checks if user is allowed to continue depending on limit
func (rl *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {

	if rl.expire != 0 {
		ttl := rl.cmdable.PTTL(ctx, key).Val()
		if ttl.Milliseconds() < 10 {
			rl.cmdable.Del(ctx, key)
			return true, nil
		}
	}

	count, err := rl.cmdable.Get(ctx, key).Int()

	if err == redis.Nil {
		return true, nil
	}

	if err != nil {
		return true, err
	}

	return count < rl.limit, nil
}

// Seen increments number of actions performed by the user
func (rl *RedisLimiter) Seen(ctx context.Context, key string) error {
	// key := rl.getKey(identifier)

	if rl.cmdable.Exists(ctx, key).Val() == 0 {
		return rl.cmdable.Set(ctx, key, 1, rl.expire).Err()
	}

	return rl.cmdable.Incr(ctx, key).Err()
}
