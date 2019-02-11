package bot

import "strings"

func (bot *Bot) handleCommand(roomID, text string) {
	parts := strings.Split(text, " ")
	switch parts[0] {
	case "!help":
		bot.client.SendText(roomID, helpText)
	case "!start":
		bot.start(roomID)
	case "!save":
		if bot.sessMan.HasSession(roomID) {
			sess, err := bot.sessMan.GetOrCreateSession(roomID)
			if err != nil {
				bot.sendAndLogError(roomID, err)
				return
			}
			sess.Close()
		}
	}
}

func (bot *Bot) handleGame(roomID, text string) {
	err := bot.start(roomID)
	if err != nil {
		bot.sendAndLogError(roomID, err)
	}

	sess, err := bot.sessMan.GetOrCreateSession(roomID)
	if err != nil {
		bot.sendAndLogError(roomID, err)
		return
	}

	err = sess.Writeln(text)
	if err != nil {
		bot.sendAndLogError(roomID, err)
	}
}
