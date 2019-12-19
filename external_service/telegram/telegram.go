package telegram

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	ce "tgj-bot/custom_errors"

	"golang.org/x/net/proxy"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TgConfig struct {
	Token         string `json:"token"`
	UpdateTimeout int    `json:"update_timeout"`
	Proxy         string `json:"proxy"`
	ChatID        int64  `json:"chat_id"`
}

type Client struct {
	Bot     *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel
	ChatID  int64
}

func RunBot(cfg TgConfig) (tgClient Client, err error) {
	c, err := initHTTPClient(cfg.Proxy)
	if err != nil {
		return
	}
	tgClient.Bot, err = tgbotapi.NewBotAPIWithClient(cfg.Token, c)
	if err != nil {
		return tgClient, errors.New("Bot connect err: " + err.Error())
	}
	tgClient.Bot.Debug = true
	log.Printf("Authorized on account %s", tgClient.Bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.UpdateTimeout

	tgClient.Updates, err = tgClient.Bot.GetUpdatesChan(u)
	if err != nil {
		return tgClient, errors.New("Update channel err: " + err.Error())
	}
	tgClient.ChatID = cfg.ChatID

	return
}

func (c *Client) SendMessage(msg string) error {
	if m, err := c.Bot.Send(tgbotapi.NewMessage(c.ChatID, msg)); err != nil {
		log.Printf("Couldn't send message '%v': %v", m, err)
		return err
	}
	return nil
}

func initHTTPClient(proxyRaw string) (*http.Client, error) {
	client := new(http.Client)

	if proxyRaw != "" {
		proxyUrl, err := url.Parse(proxyRaw)
		if err != nil {
			return nil, ce.WrapWithLog(err, "invalid proxy")
		}
		switch strings.ToLower(proxyUrl.Scheme) {
		case "http", "https":
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
			client.Transport = transport
		case "socks5", "socks5h":
			dialer, err := proxy.FromURL(proxyUrl, proxy.Direct)
			if err != nil {
				return nil, ce.WrapWithLog(err, "cannot init socks proxy")
			}
			transport := &http.Transport{
				Dial: dialer.Dial,
			}
			client.Transport = transport
		default:
			return nil, ce.WrapWithLog(fmt.Errorf("invalid proxy type: %s, supported: http or socks5", proxyUrl.Scheme), "")
		}
	}

	return client, nil
}
