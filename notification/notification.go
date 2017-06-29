package notification

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/sphiecoh/apimonitor/conf"
)

type slackMessage struct {
	Channel   string `json:"channel"`
	Username  string `json:"username"`
	IconUrl   string `json:"icon_url"`
	IconEmoji string `json:"icon_emoji"`
	Text      string `json:"text,omitempty"`
}

func NotifySlack(message, subject string, config *conf.Config) error {
	payload := slackMessage{
		Channel:   config.SlackChannel,
		Username:  config.SlackUser,
		Text:      message,
		IconEmoji: ":ghost:",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapf(err, "Marshalling slack message failed")
	}
	b := bytes.NewBuffer(data)

	resp, err := http.Post(config.SlackURL, "application/json", b)
	if err != nil {
		return errors.Wrapf(err, "Sending data to slack failed")
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("Sending data to slack failed %v", string(body))
	}
	logrus.Infof("Send message to slack channel %s", config.SlackChannel)

	return nil
}
