package app

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	db "tgj-bot/externalService/database"
	gl "tgj-bot/externalService/gitlab"
	tg "tgj-bot/externalService/telegram"
	"tgj-bot/models"

	"github.com/davecgh/go-spew/spew"
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

const urlPattern = `#([a-z]([a-z]|\d|\+|-|\.)*):(\/\/(((([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:)*@)?((\[(|(v[\da-f]{1,}\.(([a-z]|\d|-|\.|_|~)|[!\$&'\(\)\*\+,;=]|:)+))\])|((\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5]))|(([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=])*)(:\d*)?)(\/(([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)*)*|(\/((([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)+(\/(([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)*)*)?)|((([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)+(\/(([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)*)*)|((([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)){0})(\?((([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)|[\xE000-\xF8FF]|\/|\?)*)?(\#((([a-z]|\d|-|\.|_|~|[\x00A0-\xD7FF\xF900-\xFDCF\xFDF0-\xFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)|\/|\?)*)?#iS`

func (a *App) Serve() (err error) {
	for update := range a.Telegram.Updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch command(update.Message.Command()) {
			case helpCmd:
				msg.Text = helpHandler()
			case registerCmd:
				msg.Text, err = a.registerHandler(update)
			case inactiveCmd:
				msg.Text, err = a.isActiveHandler(update, false)
			case activeCmd:
				msg.Text, err = a.isActiveHandler(update, true)
			case mrCmd:

			default:
				msg.Text = helpHandler()
			}

			if err != nil {
				msg.Text = err.Error()
			}

			if _, err := a.Telegram.Bot.Send(msg); err != nil {
				log.Printf("Couldn't send message '%s': %v", msg.Text, err)
			}
		}
	}
	return
}

func helpHandler() string {
	return fmt.Sprint("/register gitlab_id [role=dev]\n" + "/mr merge_request_url\n" +
		"/inactive [@username]\n" + "/active [@username]\n" + "/mr url")
}

func (a *App) registerHandler(update tgbotapi.Update) (msg string, err error) {
	argsStr := update.Message.CommandArguments()
	if argsStr == "" {
		err = errors.New("command require two arguments. For more information use /help")
		return
	}
	args := strings.Split(strings.ToLower(argsStr), " ")

	user := models.User{
		UserBrief: models.UserBrief{
			TelegramID: strconv.Itoa(update.Message.From.ID),
			Role:       models.Developer,
		},
		GitlabID:  args[0],
		JiraID:    "",
		IsActive:  true,
	}
	if len(args) == 2 {
		role := models.Role(args[1])
		if models.IsValidRole(role) {
			user.Role = role
		} else {
			return msg, errors.New(fmt.Sprintf("second parameter (role) must be equal one of %", models.ValidRoles))
		}
	}

	if err = a.DB.SaveUser(user); err != nil {
		return msg, err
	}

	return "Successfully created!", nil

}

func (a *App) isActiveHandler(update tgbotapi.Update, isActive bool) (msg string, err error) {
	argsStr := update.Message.CommandArguments()
	var telegramID string
	if argsStr == "" {
		telegramID = strconv.Itoa(update.Message.From.ID)
	} else {
		spew.Dump(update)
		args := strings.Split(strings.ToLower(argsStr), " ")
		// что хранится в @username или как выглядит json
		// /activate @username
		telegramID = args[0]
	}

	if err = a.DB.ChangeIsActiveUser(telegramID, isActive); err != nil {
		return msg, err
	}

	return "Success!", nil
}

func (a *App) mrHandler(update tgbotapi.Update) (msg string, err error) {
	argsStr := update.Message.CommandArguments()
	if argsStr == "" {
		err = errors.New("command require one argument. For more information use /help")
		return
	}
	args := strings.Split(strings.ToLower(argsStr), " ")

	// работает ли проверка url !?
	isUrl, err := regexp.MatchString(urlPattern, args[0])
	if err != nil {
		log.Printf("compiling regexp for url failed: %v", err)
		return msg, errors.New("compiling regexp for url failed")
	}
	if !isUrl {
		return msg, errors.New("command parameter is not url")
	}

	users, err := a.DB.GetUsersWithPayload(strconv.Itoa(update.Message.From.ID))
	if err != nil {
		log.Printf("getting users failed: %v", err)
		return msg, errors.New("getting users failed")
	}

	reviewParty := getParticipants(users, a.Config.Rp)

	// сохранить в таблицу mrs и reviews
	// получить имена для каждого пользователя
	// сформировать ответное сообщение

	return
}

func getParticipants(users models.UsersPayload, cfg ReviewParty) (rp models.UsersPayload) {
	devs := users.GetN(cfg.DevNum, models.Developer)
	leads := users.GetN(cfg.LeadNum, models.Lead)

	return append(devs, leads...)
}

