package bot

import "strings"

func (bot *Bot) handleCommand(roomID, text string) {
	parts := strings.Split(text, " ")
	switch parts[0] {
	case "!help":
		bot.client.SendText(roomID, helpText)
	case "!start":
		bot.start(roomID)
	}
}

func (bot *Bot) handleGame(roomID, text string) {

}
