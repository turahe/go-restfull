package database

import (
	"testing"

	"go-rest/internal/config"

	"github.com/stretchr/testify/require"
)

func TestConnectRedis_Unreachable_Fails(t *testing.T) {
	// Use an address that won't have Redis listening (connection refused or timeout)
	cfg := config.Config{
		RedisAddr:     "127.0.0.1:16379",
		RedisPassword: "",
		RedisDB:       0,
	}
	rdb, err := ConnectRedis(cfg)
	require.Error(t, err)
	require.Nil(t, rdb)
}

