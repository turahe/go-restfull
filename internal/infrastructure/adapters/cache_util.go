// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// cacheSetJSON stores a JSON-serialized value in Redis cache with a specified TTL.
// This function is a utility for caching operations that need to serialize Go structs
// to JSON before storing them in Redis. If the Redis client is nil, the function
// returns silently (useful for environments where caching is optional).
//
// Parameters:
//   - ctx: Context for the Redis operation
//   - rdb: Redis client interface for cache operations
//   - key: Cache key to store the value under
//   - value: Value to serialize and cache (must be JSON-serializable)
//   - ttl: Time-to-live duration for the cached value
//
// Returns:
//   - error: Any error that occurred during serialization or Redis operation
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

// cacheGetJSON retrieves and deserializes a JSON value from Redis cache.
// This function attempts to retrieve a cached value by key and deserialize it
// into the provided destination pointer. If the key doesn't exist, it returns
// false without error. If the Redis client is nil, it returns false (cache miss).
//
// Type Parameters:
//   - T: The type to deserialize the cached value into
//
// Parameters:
//   - ctx: Context for the Redis operation
//   - rdb: Redis client interface for cache operations
//   - key: Cache key to retrieve the value from
//   - dest: Pointer to the destination variable for deserialized data
//
// Returns:
//   - bool: True if the value was found and deserialized successfully, false otherwise
//   - error: Any error that occurred during Redis operation or deserialization
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

// cacheDelete removes one or more keys from Redis cache.
// This function deletes multiple cache keys in a single operation. If the Redis
// client is nil or no keys are provided, the function returns silently.
//
// Parameters:
//   - ctx: Context for the Redis operation
//   - rdb: Redis client interface for cache operations
//   - keys: Variable number of cache keys to delete
//
// Returns:
//   - error: Any error that occurred during the Redis delete operation
func cacheDelete(ctx context.Context, rdb redis.Cmdable, keys ...string) error {
	if rdb == nil || len(keys) == 0 {
		return nil
	}
	return rdb.Del(ctx, keys...).Err()
}

// cacheDeleteByPattern removes cache keys that match a specific pattern.
// This function uses Redis SCAN to find keys matching the pattern and deletes
// them in batches. It's useful for clearing related cache entries without
// knowing the exact key names. The function handles pagination automatically
// and processes keys in batches of 100 for efficiency.
//
// Parameters:
//   - ctx: Context for the Redis operation
//   - rdb: Redis client interface for cache operations
//   - pattern: Redis pattern to match keys (e.g., "user:*" for all user-related keys)
//
// Returns:
//   - error: Any error that occurred during the Redis scan or delete operations
func cacheDeleteByPattern(ctx context.Context, rdb redis.Cmdable, pattern string) error {
	if rdb == nil {
		return nil
	}
	var cursor uint64
	for {
		// Scan for keys matching the pattern, processing 100 keys at a time
		keys, next, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		// Delete found keys if any exist
		if len(keys) > 0 {
			if err := rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		// Break when scan is complete (cursor returns to 0)
		if cursor == 0 {
			break
		}
	}
	return nil
}
