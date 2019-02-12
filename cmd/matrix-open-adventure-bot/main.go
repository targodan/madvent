package main

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/targodan/madvent/bot"
	"github.com/targodan/madvent/session"
)

func main() {
	log.SetLevel(log.DebugLevel)

	botConfig := &bot.Config{
		ServerURL:   "https://targodan.de",
		UserID:      "@advent:targodan.de",
		AccessToken: "",
		SyncTimeout: 500 * time.Millisecond,
	}
	manConfig := &session.Config{
		SessionTimeout: 30 * time.Minute,
		SavePath:       "/tmp/advent",
		ExecutablePath: "/usr/bin/advent",
	}

	man := session.NewManager(manConfig)
	bot, err := bot.New(man, botConfig)
	if err != nil {
		log.Panic(err)
	}

	bot.Run()
}
