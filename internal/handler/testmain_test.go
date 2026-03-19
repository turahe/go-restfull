package handler

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// Gin mode is a global mutable variable; set it once for the package
	// to avoid races when tests run in parallel with -race.
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
