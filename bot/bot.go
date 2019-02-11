package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/matrix-org/gomatrix"
	log "github.com/sirupsen/logrus"
	"github.com/targodan/madvent/adventure"
	"github.com/targodan/madvent/session"
)

const commandChar = '!'

type Config struct {
	ServerURL   string
	UserID      string
	AccessToken string
	SyncTimeout time.Duration
}

type Bot struct {
	config *Config

	client  *gomatrix.Client
	sessMan *session.Manager

	syncCtx    context.Context
	syncCancel func()
}

func New(sessMan *session.Manager, config *Config) (bot *Bot, err error) {
	bot = &Bot{
		sessMan: sessMan,
		config:  config,
	}
	bot.syncCtx, bot.syncCancel = context.WithCancel(context.Background())
	bot.client, err = gomatrix.NewClient(config.ServerURL, config.UserID, config.AccessToken)
	if err != nil {
		return nil, err
	}
	return bot, nil
}

func (bot *Bot) recvMessage(ev *gomatrix.Event) {
	text, ok := ev.Body()
	if !ok {
		return
	}
	text = strings.Trim(text, " \t\r\n")

	if text[0] == commandChar {
		bot.handleCommand(ev.RoomID, text)
	} else {
		// TODO: Start goroutine here?
		bot.handleGame(ev.RoomID, text)
	}
}

func (bot *Bot) recvRoomMember(ev *gomatrix.Event) {
	if ev.Content["membership"] == "join" {
		_, err := bot.client.JoinRoom(ev.RoomID, "", nil)
		if err != nil {
			log.Error("ERROR: Could not join room " + ev.RoomID + " although invited.")
		}
		bot.sendWelcomeMessage(ev.RoomID)
	}
}

func (bot *Bot) sendWelcomeMessage(roomID string) {
	bot.client.SendText(roomID, fmt.Sprintf(welcomeText, int(bot.config.SyncTimeout.Minutes())))
}

func (bot *Bot) Run() {
	syncer := bot.client.Syncer.(*gomatrix.DefaultSyncer)
	syncer.OnEventType("m.room.message", bot.recvMessage)
	syncer.OnEventType("m.room.member", bot.recvRoomMember)

	for {
		err := bot.client.Sync()
		if err != nil {
			log.Fatal(err)
		}
		select {
		case <-time.After(bot.config.SyncTimeout):
		case <-bot.syncCtx.Done():
			break
		}
	}
}

func filterOutput(text string) string {
	if len(text) > 2 {
		text = text[:len(text)-len(adventure.UserInteractionDelimiter)]
	}
	return strings.Trim(text, " \t\r\n")
}

func outputSender(client *gomatrix.Client, roomID string, output <-chan string) {
	for text := range output {
		_, err := client.SendText(roomID, filterOutput(text))
		if err != nil {
			log.Error(err)
		}
	}
}

func (bot *Bot) start(roomID string) error {
	if !bot.sessMan.HasSession(roomID) {
		sess, err := bot.sessMan.GetOrCreateSession(roomID)
		if err != nil {
			bot.sendAndLogError(roomID, err)
			return err
		}

		go outputSender(bot.client, roomID, sess.Output())
	}
	return nil
}

func (bot *Bot) sendAndLogError(roomID string, err error) {
	bot.client.SendText(roomID, err.Error())
	log.WithField("room", roomID).Error(err)
}

func (bot *Bot) Close(ctx context.Context) {
	bot.client.StopSync()
	bot.syncCancel()
}
