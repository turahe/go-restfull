package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"webapi/config"
	"webapi/internal/http/validation"
	"webapi/internal/logger"
	"webapi/internal/router"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: "serve", Title: "Serve:"})
	rootCmd.AddCommand(serveAPICmd)
}

var serveAPICmd = &cobra.Command{
	Use:     "server",
	Short:   "Start the restFull API",
	GroupID: "serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setup all the required dependencies
		SetupAll()

		// Create controllers router
		r := router.NewFiberRouter()

		// Create validator instance
		validation.InitValidator()

		// Get port from config
		port := config.GetConfig().HttpServer.Port

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		localIP, _ := getLocalIP()
		go func() {

			logger.Log.Info(fmt.Sprintf("Starting server on port %d", port))
			logger.Log.Info(fmt.Sprintf("Local: http://localhost:%d", port))
			logger.Log.Info(fmt.Sprintf("Network: http://%s:%d", localIP, port))

			// Log Swagger URL if configured
			swaggerURL := config.GetConfig().HttpServer.SwaggerURL
			if swaggerURL != "" {
				logger.Log.Info(fmt.Sprintf("Swagger Documentation: %s", swaggerURL))
			}

			logger.Log.Info("waiting for requests...")

			if err := r.Listen(fmt.Sprintf(":%d", port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Log.Fatal(fmt.Sprintf("listen: %s\n", err))
			}
		}()

		<-ctx.Done()
		stop()
		fmt.Println("\nShutting down gracefully, press Ctrl+C again to force")

		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := r.ShutdownWithTimeout(5 * time.Second); err != nil {
			fmt.Println(err)
		}

		return nil
	},
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("local IP not found")
}
