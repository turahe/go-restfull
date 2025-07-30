package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/db/rdb"
)

// Set sets a key-value pair with an expiration time.
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Skip cache operation if Redis is not available
		return nil
	}

	key = rdb.AddPrefix(key)
	err := redisClient.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Get retrieves the value of a key from Redis.
func Get(ctx context.Context, key string) (string, error) {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Return empty string if Redis is not available
		return "", fmt.Errorf("redis not available")
	}

	key = rdb.AddPrefix(key)
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Pull retrieves the value of a key from Redis and then deletes the key-value pair.
func Pull(ctx context.Context, key string) (string, error) {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Return empty string if Redis is not available
		return "", fmt.Errorf("redis not available")
	}

	key = rdb.AddPrefix(key)
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	_, delErr := redisClient.Del(ctx, key).Result()
	if delErr != nil {
		return "", fmt.Errorf("failed to delete key %s: %w", key, delErr)
	}

	return val, nil
}

// Forever sets the value of a key without an expiration time.
func SetForever(ctx context.Context, key string, value interface{}) error {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Skip cache operation if Redis is not available
		return nil
	}

	key = rdb.AddPrefix(key)
	err := redisClient.Set(ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s forever: %w", key, err)
	}
	return nil
}

// Delete the key-value pair from Redis.
func Remove(ctx context.Context, key string) error {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Skip cache operation if Redis is not available
		return nil
	}

	key = rdb.AddPrefix(key)
	_, err := redisClient.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to forget key %s: %w", key, err)
	}
	return nil
}

// Remove all keys from the current database.
func Flush(ctx context.Context) error {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Skip cache operation if Redis is not available
		return nil
	}

	key := rdb.AddPrefix("*")
	_, err := redisClient.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

// Increment increases the integer value of a key by the given increment.
// If the key does not exist, it is set to 0 before performing the operation.
func Increment(ctx context.Context, key string, increment int64) (int64, error) {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Return 0 if Redis is not available
		return 0, fmt.Errorf("redis not available")
	}

	key = rdb.AddPrefix(key)
	val, err := redisClient.IncrBy(ctx, key, increment).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, increment, err)
	}
	return val, nil
}

// Decrement decreases the integer value of a key by the given decrement.
// If the key does not exist, it is set to 0 before performing the operation.
func Decrement(ctx context.Context, key string, decrement int64) (int64, error) {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Return 0 if Redis is not available
		return 0, fmt.Errorf("redis not available")
	}

	key = rdb.AddPrefix(key)
	val, err := redisClient.DecrBy(ctx, key, decrement).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s by %d: %w", key, decrement, err)
	}
	return val, nil
}

func Remember(ctx context.Context, key string, duration time.Duration, fetchFunc func() ([]byte, error)) ([]byte, error) {
	value, err := Get(ctx, key)
	if err == nil {
		return []byte(value), nil
	}

	data, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	err = Set(ctx, key, data, duration)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func RememberForever(ctx context.Context, key string, fetchFunc func() ([]byte, error)) ([]byte, error) {
	return Remember(ctx, key, 0, fetchFunc)
}

// SetJSON sets a JSON-encoded value with expiration
func SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return Set(ctx, key, string(jsonData), expiration)
}

// GetJSON retrieves and unmarshals a JSON value
func GetJSON(ctx context.Context, key string, dest interface{}) error {
	jsonStr, err := Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonStr), dest)
}

// RememberJSON remembers a JSON value with automatic serialization/deserialization
func RememberJSON(ctx context.Context, key string, duration time.Duration, dest interface{}, fetchFunc func() (interface{}, error)) error {
	err := GetJSON(ctx, key, dest)
	if err == nil {
		return nil
	}

	data, err := fetchFunc()
	if err != nil {
		return err
	}

	err = SetJSON(ctx, key, data, duration)
	if err != nil {
		return err
	}

	// Update dest with the fetched data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal fetched data: %w", err)
	}
	return json.Unmarshal(jsonData, dest)
}

// InvalidatePattern removes all keys matching the given pattern
func InvalidatePattern(ctx context.Context, pattern string) error {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Skip cache operation if Redis is not available
		return nil
	}

	pattern = rdb.AddPrefix(pattern)
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		_, err = redisClient.Del(ctx, keys...).Result()
		if err != nil {
			return fmt.Errorf("failed to delete keys for pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// Exists checks if a key exists
func Exists(ctx context.Context, key string) (bool, error) {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Return false if Redis is not available
		return false, fmt.Errorf("redis not available")
	}

	key = rdb.AddPrefix(key)
	exists, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if key %s exists: %w", key, err)
	}
	return exists > 0, nil
}

// TTL gets the time to live of a key
func TTL(ctx context.Context, key string) (time.Duration, error) {
	redisClient := rdb.GetRedisClient()
	if redisClient == nil {
		// Return 0 if Redis is not available
		return 0, fmt.Errorf("redis not available")
	}

	key = rdb.AddPrefix(key)
	ttl, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return ttl, nil
}

// Default cache durations
const (
	DefaultCacheDuration   = 15 * time.Minute
	ShortCacheDuration     = 5 * time.Minute
	LongCacheDuration      = 1 * time.Hour
	VeryLongCacheDuration  = 24 * time.Hour
	PermanentCacheDuration = 0 // No expiration
)
