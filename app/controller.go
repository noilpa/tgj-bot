package app

import (
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
	db "tgj-bot/externalService/database"
	gl "tgj-bot/externalService/gitlab"
	tg "tgj-bot/externalService/telegram"
	"tgj-bot/utils"
)

type Config struct {
	Tg tg.TgConfig     `json:"telegram"`
	Gl gl.GitlabConfig `json:"gitlab"`
	Db db.DbConfig     `json:"database"`
}

type App struct {
	Telegram tg.Client
	Gitlab   gl.Client
	DB       db.Client
}

func (a *App) Serve(cfg Config) (err error) {

	u := tgbotapi.NewUpdate(cfg.Tg.UpdateOffset)
	u.Timeout = cfg.Tg.UpdateTimeout

	updates, err := a.Telegram.Bot.GetUpdatesChan(u)
	if err != nil {
		return errors.New("Update channel err: " + err.Error())
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				msg.Text = help()
			case "register":
				msg.Text, err = a.register(update)
				if err != nil {
					msg.Text = err.Error()
				}
			default:
				msg.Text = help()
			}
			if _, err := a.Telegram.Bot.Send(msg); err != nil {
				log.Printf("Couldn't send message '%s': %v", msg.Text, err)
			}
		}
	}
	return
}

func help() string {
	return fmt.Sprint("/register gitlab_id [role=dev]\n")
}

func (a *App) register(update tgbotapi.Update) (msg string, err error) {
	argsStr := update.Message.CommandArguments()
	if argsStr == "" {
		err = errors.New("Command require two arguments. For more information use /help")
		return
	}
	args := strings.Split(strings.ToLower(argsStr), " ")

	user := utils.User{
		TelegramID: strconv.Itoa(update.Message.From.ID),
		GitlabID:   args[0],
		Role:       "dev",
		IsActive:   true,
	}
	if len(args) == 2 {
		user.Role = args[1]
	}

	if err = a.DB.SaveUser(user); err != nil {
		return msg, err
	}

	return "Successfully created", nil

}


