package scheduler

import (
	"fmt"
	"time"
	_ "time/tzdata"

	"github.com/turahe/go-restfull/pkg/logger"

	"github.com/go-co-op/gocron"
	"github.com/turahe/go-restfull/config"
)

var Timezone = time.Now().Location()

func Start() {
	if config.GetConfig().Scheduler.Timezone != "" {
		Timezone, _ = time.LoadLocation(config.GetConfig().Scheduler.Timezone)
	}

	s := gocron.NewScheduler(Timezone)
	s.SingletonModeAll()

	for _, schedule := range config.GetConfig().Schedules {
		if schedule.IsEnabled {
			switch schedule.Job {
			case "DoSomeThing":
				task, err := s.CronWithSeconds(schedule.Cron).Do(func() {
					// j := job.NewJobContext()
					// j.DoSomeThing()
				})

				if err != nil {
					fmt.Printf("Failed to schedule SyncAll job: %v", err)
					continue
				}

				// Set up event listeners
				task.SetEventListeners(func() {
					fmt.Println("DoSomeThing Job started -- round: ", task.RunCount())
				}, func() {
					time.Sleep(1 * time.Second)

					// Print next run time in both utc and asia/jakarta timezone
					asiaBangkok, _ := time.LoadLocation("Asia/Jakarta")
					fmt.Printf("\nNext run: %s / %s\n", task.NextRun().UTC().String(), task.NextRun().In(asiaBangkok).String())

				})
			case "DatabaseBackup":
				// TODO: Implement backup using messaging system
				logger.Log.Info("Database backup job not implemented - use messaging system instead")
			}
		}
	}

	fmt.Printf("Total jobs: %d jobs scheduled to run\n", len(s.Jobs()))
	fmt.Printf("Timezone: %s\n", s.Location().String())
	fmt.Println("Starting scheduler... (press Ctrl+C to quit)")

	s.StartImmediately()
	s.StartBlocking()
}
