package cronjob

import (
	"github.com/turahe/go-restfull/pkg/logger"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

var scheduler gocron.Scheduler
var activeTasks map[string]string

func init() {
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Log.Error("cronjob init error", zap.Error(err))
	}

	scheduler = s
	activeTasks = make(map[string]string)

	go Start()
}

func AddJob(jobDefinition gocron.JobDefinition, taskFunc func(), jobOption gocron.JobOption) (string, error) {
	j, err := scheduler.NewJob(
		jobDefinition,
		gocron.NewTask(taskFunc),
		jobOption,
	)
	if err != nil {
		logger.Log.Error("cronjob add job error", zap.Error(err))
		return "", err
	}

	return j.ID().String(), nil
}

func Start() {
	scheduler.Start()
}

func Shutdown() error {
	return scheduler.Shutdown()
}

// MonitorDatabaseTaskChange monitors database for task changes
// This is a placeholder function - you'll need to implement the actual database logic
func MonitorDatabaseTaskChange() {
	_, err := AddJob(gocron.DurationJob(1*time.Minute), func() {
		// TODO: Implement database monitoring logic
		// For now, this is a placeholder that logs the monitoring activity
		logger.Log.Info("Monitoring database for task changes")

		// Example structure for when you implement the actual database logic:
		// db := database.GetDB()
		// var resultOfActive, resultOfInactive []entities.SystemTask
		// db.Model(&entities.SystemTask{}).Where("task_status = ?", "active").Scan(&resultOfActive)
		// db.Model(&entities.SystemTask{}).Where("task_status = ?", "inactive").Scan(&resultOfInactive)

		for taskID, jobID := range activeTasks {
			logger.Log.Info("Active task", zap.String("task_id", taskID), zap.String("job_id", jobID))
		}

	}, gocron.WithName("system"))

	if err != nil {
		logger.Log.Error("MonitorDatabaseTaskChange error", zap.Error(err))
	}
}
