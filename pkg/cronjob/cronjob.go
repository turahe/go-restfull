package cronjob

import (
	"github.com/turahe/go-restfull/pkg/logger"

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
