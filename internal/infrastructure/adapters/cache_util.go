package adapters

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

func cacheSetJSON(ctx context.Context, rdb redis.Cmdable, key string, value any, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, data, ttl).Err()
}

func cacheGetJSON[T any](ctx context.Context, rdb redis.Cmdable, key string, dest *T) (bool, error) {
	if rdb == nil {
		return false, nil
	}
	b, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return false, err
	}
	return true, nil
}

func cacheDelete(ctx context.Context, rdb redis.Cmdable, keys ...string) error {
	if rdb == nil || len(keys) == 0 {
		return nil
	}
	return rdb.Del(ctx, keys...).Err()
}

func cacheDeleteByPattern(ctx context.Context, rdb redis.Cmdable, pattern string) error {
	if rdb == nil {
		return nil
	}
	var cursor uint64
	for {
		keys, next, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
