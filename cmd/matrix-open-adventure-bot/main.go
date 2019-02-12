package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/targodan/madvent/bot"
	"github.com/targodan/madvent/session"
)

type config struct {
	Logger         *loggerConfig   `yaml:"logger"`
	Bot            *bot.Config     `yaml:"bot"`
	SessionManager *session.Config `yaml:"sessionManager"`
}

type loggerConfig struct {
	Level string `yaml:"level"`
}

func (cfg *loggerConfig) GetLevel() log.Level {
	switch cfg.Level {
	case "TRACE":
		return log.TraceLevel
	case "DEBUG":
		return log.DebugLevel
	case "INFO":
		return log.InfoLevel
	case "WARNING":
		return log.WarnLevel
	case "ERROR":
		return log.ErrorLevel
	case "FATAL":
		return log.FatalLevel
	case "PANIC":
		return log.PanicLevel
	}
	panic("Invalid logger level \"" + cfg.Level + "\"")
}

func loadConfig(filename string) (*config, error) {
	cfgData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := new(config)
	yaml.Unmarshal(cfgData, cfg)
	return cfg, nil
}

func main() {
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Panic(err)
	}

	log.SetLevel(cfg.Logger.GetLevel())

	man := session.NewManager(cfg.SessionManager)
	bot, err := bot.New(man, cfg.Bot)
	if err != nil {
		log.Panic(err)
	}

	bot.Run()
}
