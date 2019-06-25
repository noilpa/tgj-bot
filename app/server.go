package app

import (
	"log"
	"strconv"

	ce "tgj-bot/customErrors"
	db "tgj-bot/externalService/database"
	gl "tgj-bot/externalService/gitlab"
	tg "tgj-bot/externalService/telegram"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	Tg tg.TgConfig     `json:"telegram"`
	Gl gl.GitlabConfig `json:"gitlab"`
	Db db.DbConfig     `json:"database"`
	Rp ReviewParty     `json:"review_party"`
}

type ReviewParty struct {
	LeadNum int `json:"lead"`
	DevNum  int `json:"dev"`
}

type App struct {
	Telegram tg.Client
	Gitlab   gl.Client
	DB       db.Client
	Config   Config
}

type command string

const (
	helpCmd     = command("help")
	registerCmd = command("register")
	inactiveCmd = command("inactive")
	activeCmd   = command("active")
	mrCmd       = command("mr")
)

const success  = "Success!"

func (a *App) Serve() (err error) {
	for update := range a.Telegram.Updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Chat != nil {

		}
		if update.Message.Chat.ID != a.Config.Tg.ChatID {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch command(update.Message.Command()) {
			case helpCmd:
				msg.Text = a.helpHandler()
			case registerCmd:
				msg.Text, err = a.registerHandler(update)
			case inactiveCmd:
				if err = a.isUserRegister(update.Message.From.ID); err == nil {
					msg.Text, err = a.isActiveHandler(update, false)
				}
			case activeCmd:
				if err = a.isUserRegister(update.Message.From.ID); err == nil {
					msg.Text, err = a.isActiveHandler(update, true)
				}
			case mrCmd:
				if err = a.isUserRegister(update.Message.From.ID); err == nil {
					msg.Text, err = a.mrHandler(update)
				}
			default:
				msg.Text = a.helpHandler()
			}

			if err != nil {
				log.Print(err)
				msg.Text = err.Error()

			}

			if _, err := a.Telegram.Bot.Send(msg); err != nil {
				log.Printf("Couldn't send message '%s': %v", msg.Text, err)
			}
		}
	}
	return
}

func (a *App) isUserRegister(tgID int) (err error) {
	users, err := a.DB.GetUsers()
	if err != nil {
		return ce.Wrap(err, "load user from db failed")
	}

	id := strconv.Itoa(tgID)
	for _, u := range users {
		if u.TelegramID == id {
			return
		}
	}
	return ce.ErrUserNorRegistered
}

// todo add notifications about MRs