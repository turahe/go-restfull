package database

import (
	"testing"

	"go-rest/internal/config"

	"go.uber.org/zap"
	"github.com/stretchr/testify/require"
)

func TestConnectMySQL_InvalidConfig_Fails(t *testing.T) {
	// Use an invalid/unreachable host so connection fails without needing a real MySQL
	cfg := config.Config{
		DBHost:     "127.0.0.1",
		DBPort:     "17999", // typically nothing listening
		DBUser:     "root",
		DBPassword: "",
		DBName:     "test",
	}
	db, err := ConnectMySQL(cfg, zap.NewNop())
	require.Error(t, err)
	require.Equal(t, DB{}, db)
}
