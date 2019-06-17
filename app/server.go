package app

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	ce "tgj-bot/customErrors"
	db "tgj-bot/externalService/database"
	gl "tgj-bot/externalService/gitlab"
	tg "tgj-bot/externalService/telegram"
	"tgj-bot/models"

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
				msg.Text = helpHandler()
			case registerCmd:
				msg.Text, err = a.registerHandler(update)
			case inactiveCmd:
				if err = a.userValid(update.Message.From.ID); err == nil {
					msg.Text, err = a.isActiveHandler(update, false)
				}
			case activeCmd:
				if err = a.userValid(update.Message.From.ID); err == nil {
					msg.Text, err = a.isActiveHandler(update, true)
				}
			case mrCmd:
				if err = a.userValid(update.Message.From.ID); err == nil {
					msg.Text, err = a.mrHandler(update)
				}
			default:
				msg.Text = helpHandler()
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

func (a *App) userValid(tgID int) (err error) {
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
			TelegramID:       strconv.Itoa(update.Message.From.ID),
			TelegramUsername: update.Message.From.UserName,
			Role:             models.Developer,
		},
		GitlabID: args[0],
		JiraID:   "",
		IsActive: true,
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

	return success, nil

}

func (a *App) isActiveHandler(update tgbotapi.Update, isActive bool) (msg string, err error) {
	argsStr := update.Message.CommandArguments()
	var telegramUsername string
	if argsStr == "" {
		telegramUsername = update.Message.From.UserName
	} else {
		args := strings.Split(strings.ToLower(argsStr), " ")
		telegramUsername = args[0]
	}

	u, err := a.DB.GetUserByTgUsername(telegramUsername)
	if err != nil {
		return
	}
	if u.IsActive == isActive {
		// nothing to update
		return success, nil
	}

	if err = a.DB.ChangeIsActiveUser(telegramUsername, isActive); err != nil {
		return msg, err
	}

	// перераспределить МРы деактивированного пользователя
	if !isActive {
		//reallocateMRs()
	}

	return success, nil
}

func (a *App) mrHandler(update tgbotapi.Update) (msg string, err error) {
	argsStr := update.Message.CommandArguments()
	if argsStr == "" {
		err = errors.New("command require one argument. For more information use /help")
		return
	}
	args := strings.Split(strings.ToLower(argsStr), " ")

	// можно забирать ИД МРа для gitlab
	mrUrl := args[0]
	_, err = url.Parse(mrUrl)
	if err != nil {
		return
	}

	users, err := a.DB.GetUsersWithPayload(strconv.Itoa(update.Message.From.ID))
	if err != nil {
		log.Printf("getting users failed: %v", err)
		return msg, errors.New("getting users failed")
	}

	// если комманда не набирается, но не нулевая то все ОК
	reviewParty, err := getParticipants(users, a.Config.Rp)
	if err != nil {
		return
	}
	if len(reviewParty) == 0 {
		return "", ce.ErrUsersForReviewNotFound
	}

	// сохранить в таблицу mrs и reviews
	mr, err := a.DB.SaveMR(mrUrl)
	if err != nil {
		return
	}
	review := models.Review{
		MrID:       mr.ID,
		UpdatedAt:  time.Now().Unix(),
	}

	msg = fmt.Sprintf("Time to review %v !!! ", mr.URL)
	for _, r := range reviewParty {
		review.UserID = r.ID
		if err = a.DB.SaveReview(review); err != nil {
			return
		}
		msg += msg + fmt.Sprintf("@%v ", r.TelegramUsername)
	}
	msg += msg + "your turn"

	return
}

func getParticipants(users models.UsersPayload, cfg ReviewParty) (rp models.UsersPayload, err error) {
	devs, err := users.GetN(cfg.DevNum, models.Developer)
	if err != nil {
		return nil, err
	}
	leads, err := users.GetN(cfg.LeadNum, models.Lead)
	if err != nil {
		return nil, err
	}
	return append(devs, leads...), nil
}

// todo add notifications about MRs