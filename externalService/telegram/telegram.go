package telegram

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TgConfig struct {
	Token         string `json:"token"`
	UpdateTimeout int    `json:"update_timeout"`
	UpdateOffset  int    `json:"update_offset"`
}

type Client struct {
	Bot     *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel
}

func RunBot(cfg TgConfig) (tgClient Client, err error) {
	tgClient.Bot, err = tgbotapi.NewBotAPIWithClient(cfg.Token, &http.Client{})
	if err != nil {
		return tgClient, errors.New("Bot connect err: " + err.Error())
	}
	tgClient.Bot.Debug = true
	log.Printf("Authorized on account %s", tgClient.Bot.Self.UserName)

	u := tgbotapi.NewUpdate(cfg.UpdateOffset)
	u.Timeout = cfg.UpdateTimeout

	tgClient.Updates, err = tgClient.Bot.GetUpdatesChan(u)
	if err != nil {
		return tgClient, errors.New("Update channel err: " + err.Error())
	}

	return
}
