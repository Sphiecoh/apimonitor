package schedule

import (
	"time"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/sphiecoh/apimonitor/conf"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/notification"
	"gopkg.in/robfig/cron.v2"
)

//Scheduler mantains the jobs and crons
type Scheduler struct {
	Cron   *cron.Cron
	Jobs   []*RunnerJob
	Store  *db.Store
	Config *conf.Config
}

//RunnerJob  represents the job to run by cron
type RunnerJob struct {
	db     *db.Store
	target *db.ApiTest
	Next   time.Time
	Prev   time.Time
	Config *conf.Config
}

// ToJob converts a test to a job
func ToJob(test *db.ApiTest, store *db.Store, conf *conf.Config) *RunnerJob {
	job := &RunnerJob{
		target: test,
		db:     store,
		Config: conf,
	}
	return job
}

//New creates a schedular
func New(tests []*db.ApiTest, store *db.Store, conf *conf.Config) *Scheduler {
	jobs := make([]*RunnerJob, 0)
	for _, test := range tests {
		job := &RunnerJob{
			db:     store,
			target: test,
			Config: conf,
		}
		jobs = append(jobs, job)
		logrus.Infof("Scheduling test %s", test.Name)
	}
	s := &Scheduler{
		Cron:   cron.New(),
		Jobs:   jobs,
		Store:  store,
		Config: conf,
	}
	return s
}

//Start starts the scheduler
func (s *Scheduler) Start() error {
	for _, job := range s.Jobs {
		schedule, err := cron.Parse(job.target.Cron)
		if err != nil {
			return errors.Wrapf(err, "Invalid cron %v for test %v", job.target.Cron, job.target.Name)
		}
		s.Cron.Schedule(schedule, job)
	}
	s.Cron.Start()
	logrus.Info("Started job scheduler")
	return nil
}

//Run runs the cron job
func (job RunnerJob) Run() {
	var logger = logrus.WithField("name", job.target.Name)
	logger.Infof("Running test %s", job.target.Name)
	result := job.target.Run()
	logger.WithField("status", result.Status)
	if err := job.db.SaveResult(result); err != nil {
		logger.Errorf("failed to save result %v", err)
	}

	if result.Status != 200 {
		logger.Errorf("Test %s failed", job.target.Name)
		notification.NotifySlack(result.Error, fmt.Sprintf("Test %s failed", job.target.Name), job.Config)
		return
	}
	logger.Infof("Test %s succeeded", job.target.Name)
}
