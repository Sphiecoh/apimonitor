package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/caarlos0/env"
	"github.com/pkg/errors"
	"github.com/sphiecoh/apimonitor/api"
	"github.com/sphiecoh/apimonitor/conf"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/schedule"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	config := conf.Config{}
	flag.StringVar(&config.Port, "Port", ":8009", "HTTP port to listen on")
	flag.StringVar(&config.DbPath, "DataPath", "", "db dir")
	flag.StringVar(&config.SlackURL, "SlackUrl", "https://hooks.slack.com/services/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", "Slack WebHook Url")
	flag.StringVar(&config.SlackChannel, "SlackChannel", "#general", "Slack channel")
	flag.StringVar(&config.SlackUser, "slackuser", "user", "Slack username")
	flag.StringVar(&config.ConfigPath, "conf", "./tests", "Dir with tests")
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

	if config.ConfigPath != "" {
		t, err := loadTests(config.ConfigPath)
		if err != nil {
			logrus.Infof("Failed to read tests from %s ,error %v", config.ConfigPath, err)
		}
		for index := 0; index < len(t); index++ {
			test := t[index]
			test.ID = db.GenerateID()
			d, _ := json.Marshal(test)
			store.Put(test.Name, store.TestBucket, d)
		}

	}

	//Fetch tests from database
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
func loadTests(rootDir string) ([]*db.ApiTest, error) {
	if _, err := os.Stat(rootDir); err != nil {
		return nil, err
	}
	var tests = make([]*db.ApiTest, 0)
	var err error
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		//Skip directories and file not a json
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}
		var buf []byte
		buf, err = ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "Could not read file %s", path)
		}
		var tst = make([]*db.ApiTest, 0)
		err = json.Unmarshal(buf, &tst)
		if err != nil {
			return errors.Wrapf(err, "Invalid json from file %s", path)
		}
		tests = append(tests, tst...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(tests) > 0 {
		return tests, nil
	}
	return nil, nil
}
