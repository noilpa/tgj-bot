package app

import (
	"log"

	ce "tgj-bot/customErrors"
	db "tgj-bot/external_service/database"
	gl "tgj-bot/external_service/gitlab"
	tg "tgj-bot/external_service/telegram"
)

type Config struct {
	Tg       tg.TgConfig     `json:"telegram"`
	Gl       gl.GitlabConfig `json:"gitlab"`
	Db       db.DbConfig     `json:"database"`
	Rp       ReviewParty     `json:"review_party"`
	Notifier NotifierConfig  `json:"notifier"`
}

type ReviewParty struct {
	LeadNum int `json:"lead"`
	DevNum  int `json:"dev"`
}

type NotifierConfig struct {
	IsAllow    bool  `json:"is_allow"`
	TimeHour   int   `json:"time_hour"`
	TimeMinute int   `json:"time_minute"`
	Delay      int64 `json:"delay"`
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
	a.notify()
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
				if _, err = a.isUserRegister(tgUsername); err == nil {
					err = a.isActiveHandler(update, false)
				}
			case activeCmd:
				if _, err = a.isUserRegister(tgUsername); err == nil {
					err = a.isActiveHandler(update, true)
				}
			case mrCmd:
				if _, err = a.isUserRegister(tgUsername); err == nil {
					err = a.mrHandler(update)
				}
			default:
				err = a.helpHandler()
			}

			if err != nil {
				// mb duplicate database logging
				log.Print(err)
				log.Println(a.sendTgMessage(err.Error()))
			}
		}
	}
	return
}

func (a *App) isUserRegister(tgUsername string) (int, error) {
	u, err := a.DB.GetUserByTgUsername(tgUsername)
	if err != nil {
		err = ce.WrapWithLog(err, ce.ErrUserNorRegistered.Error())
	}
	return int(u.ID), err
}
