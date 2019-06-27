package app

import (
	"log"

	ce "tgj-bot/customErrors"
	db "tgj-bot/externalService/database"
	gl "tgj-bot/externalService/gitlab"
	tg "tgj-bot/externalService/telegram"
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

const success = "Success!"

func (a *App) Serve() (err error) {
	for update := range a.Telegram.Updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Chat != nil {
			if update.Message.Chat.ID != a.Config.Tg.ChatID {
				continue
			}
		}
		tgUsername := update.Message.From.UserName
		if update.Message.IsCommand() {
			switch command(update.Message.Command()) {
			case helpCmd:
				err = a.helpHandler()
			case registerCmd:
				err = a.registerHandler(update)
			case inactiveCmd:
				if err = a.isUserRegister(tgUsername); err == nil {
					err = a.isActiveHandler(update, false)
				}
			case activeCmd:
				if err = a.isUserRegister(tgUsername); err == nil {
					err = a.isActiveHandler(update, true)
				}
			case mrCmd:
				if err = a.isUserRegister(tgUsername); err == nil {
					err = a.mrHandler(update)
				}
			default:
				err = a.helpHandler()
			}

			if err != nil {
				log.Print(err)
				// no need to handle err, we are log all that we can
				a.sendTgMessage(err.Error())
			}
		}
	}
	return
}

func (a *App) isUserRegister(tgUsername string) (err error) {
	if _, err = a.DB.GetUserByTgUsername(tgUsername); err != nil {
		err = ce.ErrUserNorRegistered
	}
	return
}
