package telegram

import (
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type TgConfig struct {
	Token         string `json:"token"`
	UpdateTimeout int    `json:"update_timeout"`
	UpdateOffset  int    `json:"update_offset"`
}

type Client struct {
	Bot *tgbotapi.BotAPI
}

func RunBot(cfg TgConfig) (tgClient Client, err error) {
	tgClient.Bot, err = tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return tgClient, errors.New("Bot connect err: " + err.Error())
	}
	tgClient.Bot.Debug = true
	log.Printf("Authorized on account %s", tgClient.Bot.Self.UserName)
	return
}
