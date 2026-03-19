package main

import (
	"go-rest/internal/app"

	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "serve",
		Short:        "Run the HTTP API server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Serve(cmd.Context())
		},
	}
}

