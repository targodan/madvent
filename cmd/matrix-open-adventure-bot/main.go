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
		AccessToken: "MDAxOWxvY2F0aW9uIHRhcmdvZGFuLmRlCjAwMTNpZGVudGlmaWVyIGtleQowMDEwY2lkIGdlbiA9IDEKMDAyNmNpZCB1c2VyX2lkID0gQGFkdmVudDp0YXJnb2Rhbi5kZQowMDE2Y2lkIHR5cGUgPSBhY2Nlc3MKMDAyMWNpZCBub25jZSA9IEdoJnlxYkMwb1ZfVWs3aEgKMDAyZnNpZ25hdHVyZSAJqz_V0_YQrrMk9rWw-gpT58hHjEhYJz3-tGFVaJawugo",
		SyncTimeout: 500 * time.Millisecond,
	}
	manConfig := &session.Config{
		SessionTimeout: 30 * time.Minute,
		SavePath:       "/tmp/advent/",
		ExecutablePath: "/home/luca/open-adventure/advent",
	}

	man := session.NewManager(manConfig)
	bot, err := bot.New(man, botConfig)
	if err != nil {
		log.Panic(err)
	}

	bot.Run()
}
