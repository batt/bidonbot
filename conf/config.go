package config

import "github.com/BurntSushi/toml"

type Global struct {
	Bot Bot `toml:"bot"`
}

type Bot struct {
	SlackToken  string `toml:"slack_token"`
	ChannelID   string `toml:"channel_id"`
	TestLogFile string `toml:"test_log_file"`
}

func Load(path string) (*Global, error) {
	c := new(Global)
	_, err := toml.DecodeFile(path, &c)
	return c, err
}
