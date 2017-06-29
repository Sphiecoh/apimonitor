package conf

//Config ..
type Config struct {
	Port         string `env:"APIMON_PORT"`
	DbPath       string `env:"APIMON_DBPath"`
	SlackChannel string `env:"APIMON_SLACK_CHANNEL"`
	SlackUser    string `env:"APIMON_SLACK_USER"`
	SlackURL     string `env:"APIMON_SLACK_URL"`
}
