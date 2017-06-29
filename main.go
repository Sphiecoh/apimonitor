package main

import (
	"flag"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/caarlos0/env"
	"github.com/sphiecoh/apimonitor/api"
	"github.com/sphiecoh/apimonitor/conf"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/schedule"
)

func main() {
	config := conf.Config{}
	flag.StringVar(&config.Port, "Port", ":8009", "HTTP port to listen on")
	flag.StringVar(&config.DbPath, "DataPath", "", "db dir")
	flag.StringVar(&config.SlackURL, "SlackUrl", "https://hooks.slack.com/services/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", "Slack WebHook Url")
	flag.StringVar(&config.SlackChannel, "SlackChannel", "#general", "Slack channel")
	flag.StringVar(&config.SlackUser, "slackuser", "user", "Slack username")
	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		logrus.Error(err)
	}

	store, err := db.NewStore(path.Join(config.DbPath, "apimonitor.db"))
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logrus.Error(err)
		}
	}()
	createerr := store.CreateBuckets()
	if createerr != nil {
		logrus.Fatal(createerr)
	}

	tests, err := store.GetAllTests()
	if err != nil {
		logrus.Fatal(err)

	}
	schedule := schedule.New(tests, store, &config)
	schedulererror := schedule.Start()
	if schedulererror != nil {
		logrus.Fatalf("Failed to start scheduler %v", schedulererror)
	}
	defer schedule.Cron.Stop()

	srv := &api.Server{
		C: &config,
		H: api.Handler{S: schedule, Store: store},
	}
	go srv.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	logrus.Infof("Shutting down %v signal received", sig)

}
