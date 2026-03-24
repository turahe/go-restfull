package main

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "go-restfull",
		Short:        "Blog REST API (server + utilities)",
		SilenceUsage: true,
	}

	serveCmd := newServeCmd()
	seedCmd := newSeedCmd()
	seedCmd.AddCommand(newSeedRBACCmd())
	seedCmd.AddCommand(newSeedSettingsCmd())

	root.AddCommand(serveCmd, seedCmd)

	// Backwards compatible: running without args starts server.
	root.RunE = serveCmd.RunE

	return root
}

func execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

