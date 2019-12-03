package app

import (
	"context"
	"log"
	"strings"
	"time"

	ce "tgj-bot/custom_errors"
	db "tgj-bot/external_service/database"
	gl "tgj-bot/external_service/gitlab"
	"tgj-bot/external_service/jira"
	tg "tgj-bot/external_service/telegram"
	"tgj-bot/models"
)

type Config struct {
	Tg       tg.TgConfig     `json:"telegram"`
	Gl       gl.GitlabConfig `json:"gitlab"`
	Db       db.DbConfig     `json:"database"`
	Rp       ReviewParty     `json:"review_party"`
	Notifier NotifierConfig  `json:"notifier"`
	Jira     jira.Config     `json:"jira"`
}

type ReviewParty struct {
	LeadNum int `json:"lead"`
	DevNum  int `json:"dev"`
}

type NotifierConfig struct {
	IsAllow    bool     `json:"is_allow"`
	TimeHour   int      `json:"time_hour"`
	TimeMinute int      `json:"time_minute"`
	Delay      int64    `json:"delay"`
	Praise     []string `json:"praise"`
	Motivate   []string `json:"motivate"`
}

type App struct {
	Telegram tg.Client
	Gitlab   gl.Client
	DB       db.Client
	Config   Config
	Jira     *jira.Jira
}

type command string

const (
	helpCmd     = command("help")
	registerCmd = command("register")
	inactiveCmd = command("inactive")
	activeCmd   = command("active")
	mrCmd       = command("mr")
)

const success = "Success! 👍"

func (a *App) Serve() (err error) {
	a.notify()
	a.runCheckUpdates()

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
		if !update.Message.IsCommand() {
			continue
		}
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
			log.Print(err)
			a.Telegram.SendMessage(err.Error())
		}
	}
	return
}

func (a *App) isUserRegister(tgUsername string) (int, error) {
	u, err := a.DB.GetUserByTgUsername(strings.ToLower(tgUsername))
	if err != nil {
		err = ce.WrapWithLog(err, ce.ErrUserNorRegistered.Error())
	}
	return int(u.ID), err
}

func (a *App) runCheckUpdates() {
	go func() {
		for _ = range time.Tick(time.Second) {
			ctx := context.Background()
			log.Println("updating mrs info from jira...")
			mrs, err := a.DB.GetOpenedMRs()
			if err != nil {
				a.logError(err)
				continue
			}

			for _, mr := range mrs {
				if err := a.updateInfoFromJira(ctx, mr); err != nil {
					a.logError(err)
				}
			}
		}
	}()
}

func (a *App) updateInfoFromJira(ctx context.Context, mr models.MR) error {
	isChanged := false
	if mr.JiraID == 0 {
		gitlabID, err := mr.GetGitlabID()
		if err != nil {
			return err
		}

		title, err := a.Gitlab.GetMrTitle(gitlabID)
		if err != nil {
			return err
		}

		mr.ExtractJiraID(title)
		isChanged = true
	}

	if mr.JiraID > 0 {
		info, err := a.Jira.LoadIssueByID(ctx, mr.JiraID)
		if err != nil {
			return err
		}
		if info != nil {
			mr.JiraPriority = info.Priority
			isChanged = true
		}
	}

	if isChanged {
		if _, err := a.DB.SaveMR(mr); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) logError(err error) {
	log.Println(err)
}
