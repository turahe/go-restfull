package rdb

import (
	"context"
	"fmt"
	"github.com/turahe/go-restfull/pkg/logger"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turahe/go-restfull/config"
	"go.uber.org/zap"
)

var rdb redis.Cmdable
var m sync.Mutex
var prefix string
var queuePrefix string

type RedisCredentials struct {
	Password string
	Database int
}

func InitRedisClient(redisConfigs []config.Redis) error {
	m.Lock()
	defer m.Unlock()

	// Prepare a list of Redis addresses
	// Prepare a list of Redis addresses and a map of their corresponding credentials
	var addrs []string
	creds := make(map[string]RedisCredentials)
	for _, redisConfig := range redisConfigs {
		addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)
		addrs = append(addrs, addr)
		creds[addr] = RedisCredentials{
			Password: redisConfig.Password,
			Database: redisConfig.Database,
		}
	}

	if len(addrs) == 1 {
		rdb = redis.NewClient(&redis.Options{
			Addr:         addrs[0],
			Password:     creds[addrs[0]].Password,
			DB:           creds[addrs[0]].Database,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		})
	} else {
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: addrs,
			NewClient: func(opt *redis.Options) *redis.Client {
				cred := creds[opt.Addr]
				opt.Password = cred.Password
				opt.DB = cred.Database
				opt.DialTimeout = 5 * time.Second
				opt.ReadTimeout = 3 * time.Second
				opt.WriteTimeout = 3 * time.Second

				return redis.NewClient(opt)
			},
		})
	}

	// Add timeout to the ping operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Set the prefix string
	// for whoever is using AddPrefix() or GetPrefix()
	prefix = config.GetConfig().App.NameSlug
	queuePrefix = config.GetConfig().App.NameSlug + "_queue"

	return nil
}

func GetRedisClient() redis.Cmdable {
	if rdb == nil {
		m.Lock()
		defer m.Unlock()

		logger.Log.Info("Initializing redis again")

		// Use a context with timeout for the entire initialization
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Run initialization in a goroutine with timeout
		done := make(chan error, 1)
		go func() {
			done <- InitRedisClient(config.GetConfig().Redis)
		}()

		select {
		case err := <-done:
			if err != nil {
				logger.Log.Error("Failed to initialize redis client", zap.Error(err))
				return nil
			}
			logger.Log.Info("redis initialized")
		case <-ctx.Done():
			logger.Log.Error("Redis initialization timed out")
			return nil
		}
	}

	return rdb
}

func AddPrefix(key string) string {
	if prefix == "" {
		m.Lock()
		defer m.Unlock()
		prefix = config.GetConfig().App.NameSlug
	}
	return fmt.Sprintf("%s_%s", prefix, key)
}

func AddQueuePrefix(key string) string {
	if queuePrefix == "" {
		m.Lock()
		defer m.Unlock()
		queuePrefix = config.GetConfig().App.NameSlug + "_queue"
	}
	return fmt.Sprintf("%s_%s", queuePrefix, key)
}

func GetPrefix() string {
	return prefix
}

func GetQueuePrefix() string {
	return queuePrefix
}
