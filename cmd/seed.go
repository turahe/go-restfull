package main

import "github.com/spf13/cobra"

func newSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seed data into the database",
	}
}

