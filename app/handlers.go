package app

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"tgj-bot/models"

	ce "tgj-bot/customErrors"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func (a *App) helpHandler() string {
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
	author, err := a.DB.GetUserByTgUsername(update.Message.From.UserName)
	if err != nil {
		return
	}
	mr := models.MR{
		URL:      mrUrl,
		AuthorID: author.ID,
	}
	mr, err = a.DB.SaveMR(mr)
	if err != nil {
		return
	}
	review := models.Review{
		MrID:      mr.ID,
		UpdatedAt: time.Now().Unix(),
	}

	msg = fmt.Sprintf("Merge request %v !!! ", mr.URL)
	for _, r := range reviewParty {
		review.UserID = r.ID
		if err = a.DB.SaveReview(review); err != nil {
			return
		}
		msg += fmt.Sprintf("@%v ", r.TelegramUsername)
	}
	msg += "review please"

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

func (a *App) updateReviews(timeout int64) error {
	//
	// пойти по всем МРам
	//    зайти в мр
	//    получить лайки и коменты
	//    обновить значения в таблице reviews
	//
	mrs, err := a.DB.GetOpenedMRs()
	if err != nil {
		return nil
	}
	for _, mr := range mrs {
		if err = a.updateMrLikes(mr.ID); err != nil {
			return err
		}
		if err = a.updateMrComments(mr.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) updateMrLikes(mID int) error {
	usersID, err := a.Gitlab.CheckMrLikes(mID)
	if err != nil {
		return err
	}
	for _, uID := range usersID {
		u, err := a.DB.GetUserByGitlabID(uID)
		if err != nil {
			return err
		}
		err = a.DB.UpdateReviewApprove(models.Review{
			MrID:       mID,
			UserID:     u.ID,
			IsApproved: true,
			UpdatedAt:  time.Now().Unix(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) updateMrComments(mID int) error {
	comments, err := a.Gitlab.CheckMrComments(mID)
	if err != nil {
		return err
	}
	for uID := range comments {
		u, err := a.DB.GetUserByGitlabID(uID)
		if err != nil {
			return err
		}
		// UpdateReviewComment()
	}
	return nil
}
