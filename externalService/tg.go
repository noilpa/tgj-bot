package externalService

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type TgConfig struct {
	Token         string `json:"token"`
	UpdateTimeout int    `json:"update_timeout"`
	UpdateOffset  int    `json:"update_offset"`
}

func RunBot(tc TgConfig) error {
	bot, err := tgbotapi.NewBotAPI(tc.Token)
	if err != nil {
		return err
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(tc.UpdateOffset)
	u.Timeout = tc.UpdateTimeout

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
	return nil
}
