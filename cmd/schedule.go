package cmd

import (
	"github.com/turahe/go-restfull/pkg/scheduler"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lnquy/cron"
	"github.com/spf13/cobra"
	"github.com/turahe/go-restfull/config"
)

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: "schedule", Title: "Schedule:"})
	rootCmd.AddCommand(
		listScheduleCommand,
		startScheduleCommand,
	)
}

var startScheduleCommand = &cobra.Command{
	Use:     "schedule:run",
	Short:   "Start schedule job",
	GroupID: "schedule",
	Run: func(_ *cobra.Command, _ []string) {
		// Setup all the required dependencies
		SetupAll()

		printScheduleList()
		scheduler.Start()
	},
}

var listScheduleCommand = &cobra.Command{
	Use:     "schedule:list",
	Short:   "List all schedule jobs",
	GroupID: "schedule",
	Run: func(_ *cobra.Command, _ []string) {
		// Setup all the required dependencies
		setUpConfig()
		setUpLogger()

		printScheduleList()

	},
}

func printScheduleList() {
	exprDesc, _ := cron.NewDescriptor()

	// Print the job list as a table in the console
	tableWriter := table.NewWriter()
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.AppendHeader(table.Row{"No.", "Job Name", "Cron Expression", "Schedule"})
	for i, schedule := range config.GetConfig().Schedules {
		desc, _ := exprDesc.ToDescription(schedule.Cron, cron.Locale_en)

		tableWriter.AppendRow(table.Row{
			i + 1,
			schedule.Job,
			schedule.Cron,
			desc,
		})

	}

	tableWriter.Render()
}
