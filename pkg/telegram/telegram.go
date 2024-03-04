package telegram

import (
	"log"

	tb "gopkg.in/telebot.v3"
)

func TgSendMessage(botToken string, msg string, chatID int64) (responce *tb.Message) {
	log.Printf("Sending to chat: %d", chatID)

	tbot, err := tb.NewBot(tb.Settings{
		Token: botToken,
	})
	if err != nil {
		log.Println(err)
	} else {
		group := tb.ChatID(chatID)
		var opts tb.SendOptions
		opts.ParseMode = tb.ModeHTML
		responce, err = tbot.Send(group, msg, &opts)
		if err != nil {
			log.Println(err)
			log.Println(msg)
		}
	}
	return responce
}
