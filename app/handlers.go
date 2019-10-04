package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"tgj-bot/models"

	ce "tgj-bot/custom_errors"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func (a *App) helpHandler() error {
	a.Telegram.SendMessage(fmt.Sprint("/register gitlab_id [role=dev]\n" + "/mr merge_request_url\n" +
		"/inactive [username]\n" + "/active [username]\n"))
	return nil
}

func (a *App) registerHandler(update tgbotapi.Update) (err error) {
	argsStr := update.Message.CommandArguments()
	if argsStr == "" {
		err = errors.New("command require two arguments. For more information use /help")
		return
	}
	args := strings.Split(strings.ToLower(argsStr), " ")

	user := models.User{
		UserBrief: models.UserBrief{
			TelegramID:       strconv.Itoa(update.Message.From.ID),
			TelegramUsername: strings.ToLower(update.Message.From.UserName),
			Role:             models.Developer,
		},
		JiraID:   "",
		IsActive: true,
	}
	if len(args) == 2 {
		role := models.Role(args[1])
		if models.IsValidRole(role) {
			user.Role = role
		} else {
			return errors.New(fmt.Sprintf("second parameter (role) must be equal one of %v", models.ValidRoles))
		}
	}

	user.GitlabID, user.GitlabName, err = a.getGitlabInfo(args[0])
	if err != nil {
		return ce.WrapWithLog(err, fmt.Sprintf("get gitlab info %v", args[0]))
	}

	if _, err = a.DB.SaveUser(user); err != nil {
		return err
	}
	a.Telegram.SendMessage(success)
	return
}

func (a *App) isActiveHandler(update tgbotapi.Update, isActive bool) (err error) {
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
		a.Telegram.SendMessage(success)
		return
	}

	if err = a.DB.ChangeIsActiveUser(telegramUsername, isActive); err != nil {
		return
	}

	// reallocate MRs for inactive user
	if !isActive {
		if err = a.reallocateUserMRs(u); err != nil {
			return
		}
	}
	a.Telegram.SendMessage(success)
	return
}

func (a *App) mrHandler(update tgbotapi.Update) (err error) {
	argsStr := update.Message.CommandArguments()
	if argsStr == "" {
		err = errors.New("command require one argument. For more information use /help")
		return
	}
	args := strings.Split(strings.ToLower(argsStr), " ")

	mrUrl := args[0]
	mrID, err := models.GetGitlabID(mrUrl)
	if err != nil {
		return
	}

	if a.isMrAlreadyExist(mrUrl) {
		return a.returnMrParty(mrUrl)
	}

	authorGitlabID, err := a.Gitlab.GetMrAuthorID(mrID)
	if err != nil {
		return
	}
	author, err := a.DB.GetUserByGitlabID(authorGitlabID)
	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
		author = models.User{UserBrief: models.UserBrief{ID: 0}}
	}

	if err = a.updateReviews(); err != nil {
		return
	}

	users, err := a.DB.GetUsersWithPayload(author.TelegramID)
	if err != nil {
		log.Printf("getting users failed: %v", err)
		return errors.New("getting users failed")
	}

	// if the party is not picked up, but not zero, then everything is OK
	reviewParty, err := getParticipants(users, a.Config.Rp)
	if err != nil {
		return
	}
	if len(reviewParty) == 0 {
		return ce.ErrUsersForReviewNotFound
	}

	mr := models.MR{
		URL:      mrUrl,
		AuthorID: &author.ID,
	}
	mr, err = a.DB.SaveMR(mr)
	if err != nil {
		return
	}

	review := models.Review{
		MrID:      mr.ID,
		UpdatedAt: time.Now().Unix(),
	}

	msg := "New merge request " + randJoyEmoji() + "\n"
	for i, r := range reviewParty {

		if r.GitlabName == "" {
			r.GitlabName, err = a.Gitlab.GetUserByID(r.GitlabID)
			if err != nil {
				return ce.WrapWithLog(err, fmt.Sprintf("MR handler fail to get gitlab name for %v", r.TelegramUsername))
			}
			if _, err = a.DB.SaveUser(models.User{UserBrief: r.UserBrief}); err != nil {
				return ce.WrapWithLog(err, "MR handler fail")
			}
		}

		if r.GitlabID == 0 {
			r.GitlabID, err = a.Gitlab.GetUserByName(r.GitlabName)
			if err != nil {
				return ce.WrapWithLog(err, fmt.Sprintf("MR handler fail to get gitlab id for %v", r.TelegramUsername))
			}
			if _, err = a.DB.SaveUser(models.User{UserBrief: r.UserBrief}); err != nil {
				return ce.WrapWithLog(err, "MR handler fail")
			}
		}

		review.UserID = r.ID
		if err = a.DB.SaveReview(review); err != nil {
			return
		}
		msg += fmt.Sprintf("%s @%v\n", pointEmoji[i%2], r.TelegramUsername)
	}

	msg += cutoff + "\n" + mrUrl

	gitlabMrID, err := models.GetGitlabID(mr.URL)
	if err != nil {
		log.Println(err)
		return err
	}
	if err = a.Gitlab.WriteReviewers(gitlabMrID, reviewParty); err != nil {
		log.Println(err)
		return
	}

	a.Telegram.SendMessage(msg)
	return
}

func (a *App) updateReviews() error {
	//
	// get all open MRI
	// 	 go to mr
	// 	 get likes and comments
	// 	 update values in the table reviews
	// close MRs with all approved reviews
	//
	mrs, err := a.DB.GetOpenedMRs()
	if err != nil {
		return nil
	}
	for _, mr := range mrs {
		mr.GitlabID, err = models.GetGitlabID(mr.URL)
		if err != nil {
			_ = ce.WrapWithLog(err, "Parse gitlab mr id")
			continue
		}
		mrIsOpen, err := a.Gitlab.MrIsOpen(mr.GitlabID)
		log.Printf("Update reviews mr_id=%d is_open=%v", mr.ID, mrIsOpen)
		if !mrIsOpen {
			_ = ce.WrapWithLog(a.DB.CloseMR(mr.ID), "close mr err")
		}
		if err = a.updateMrLikes(mr); err != nil {
			_ = ce.WrapWithLog(err, "update mr likes")
		}
		if err = a.updateMrComments(mr); err != nil {
			_ = ce.WrapWithLog(err, "update mr comments")
		}
	}

	if err = a.DB.CloseMRs(); err != nil {
		return err
	}

	return nil
}

func (a *App) updateMrLikes(mr models.MR) error {
	usersID, err := a.Gitlab.CheckMrLikes(mr.GitlabID)
	if err != nil {
		return err
	}
	log.Printf("Check mr likes user's ids: %v", usersID)
	for uID := range usersID {
		u, err := a.DB.GetUserByGitlabID(uID)
		if err != nil {
			ce.WrapWithLog(err, fmt.Sprintf("user not found by gitlab id=%d", uID))
			continue
		}

		now := time.Now().Unix()
		now = a.skipWeekends(now)

		err = a.DB.UpdateReviewApprove(models.Review{
			MrID:       mr.ID,
			UserID:     u.ID,
			IsApproved: true,
			UpdatedAt:  now,
		})
		if err != nil {
			ce.WrapWithLog(err, "Update review approve err")
			continue
		}
	}
	return nil
}

func (a *App) updateMrComments(mr models.MR) error {
	usersID, err := a.Gitlab.CheckMrComments(mr.GitlabID)
	if err != nil {
		return err
	}
	log.Printf("Check mr comments user's ids: %v", usersID)
	for uID := range usersID {
		u, err := a.DB.GetUserByGitlabID(uID)
		if err != nil {
			ce.WrapWithLog(err, fmt.Sprintf("user not found by gitlab id=%d", uID))
			continue
		}

		now := time.Now().Unix()
		now = a.skipWeekends(now)

		err = a.DB.UpdateReviewComment(models.Review{
			MrID:        mr.ID,
			UserID:      u.ID,
			IsCommented: true,
			UpdatedAt:   now,
		})
		if err != nil {
			ce.WrapWithLog(err, "Update comment approve err")
			continue
		}
	}
	return nil
}

func (a *App) reallocateUserMRs(u models.User) (err error) {
	// get list unapproved user's mrs
	//     get mr's reviewers list
	//     generate review candidate list (not mr's author is active, not already review this mr, same role)
	//     get user with min payload
	//     repeat

	mrsID, err := a.DB.GetReviewMRsByUserID(u.ID)
	if err != nil {
		return err
	}
	log.Printf("Reallocate MRs for %s: %v\n", u.TelegramUsername, mrsID)
	// continue on error in the hope of the best
	for _, mrID := range mrsID {
		user, err := a.DB.GetUserForReallocateMR(u.UserBrief, mrID)
		if err != nil {
			log.Println(ce.Wrap(err, "Reallocate MRs"))
			continue
		}
		// update review
		if err = a.DB.SaveReview(models.Review{
			MrID:      mrID,
			UserID:    user.ID,
			UpdatedAt: time.Now().Unix(),
		}); err != nil {
			log.Println(ce.Wrap(err, "Reallocate MRs"))
			continue
		}
		mr, err := a.DB.GetMrByID(mrID)
		if err != nil {
			log.Println(ce.Wrap(err, "Reallocate MRs"))
			continue
		}
		a.Telegram.SendMessage(fmt.Sprintf("New review:\n@%s\n%s\n%v", user.TelegramUsername, cutoff, mr.URL))
	}
	return nil
}

func (a *App) skipWeekends(endTime int64) (newEndTime int64) {
	// +1 day -> monday
	if time.Unix(endTime+a.Config.Notifier.Delay, 0).Weekday() == time.Sunday {
		endTime += time.Unix(24*60*60, 0).Unix()
	}

	// +2 days -> monday
	if time.Unix(endTime+a.Config.Notifier.Delay, 0).Weekday() == time.Saturday {
		endTime += time.Unix(2*24*60*60, 0).Unix()
	}
	return
}

func (a *App) isMrAlreadyExist(url string) bool {
	mr, err := a.DB.GetMRbyURL(url)
	if err != nil {
		return false
	}
	defaultValue := models.MR{}
	if mr != defaultValue {
		return true
	}
	return false
}

func (a *App) returnMrParty(url string) (err error) {
	us, err := a.DB.GetUsersByMrURL(url)
	msg := "Review party:\n"
	for i, u := range us {
		msg += fmt.Sprintf("%s %s\n", pointEmoji[i%2], u.TelegramUsername)
	}
	msg += cutoff + fmt.Sprintf("\n%s", url)
	a.Telegram.SendMessage(msg)
	return
}

func (a *App) getGitlabInfo(arg string) (int, string, error) {
	invalidArgument := errors.New(fmt.Sprintf("invalid argument %v", arg))

	id, err := strconv.Atoi(arg)
	if err != nil {
		id, err = a.Gitlab.GetUserByName(arg)
		if err != nil {
			return 0, "", invalidArgument
		}
	}
	name, err := a.Gitlab.GetUserByID(id)
	if err != nil {
		return 0, "", invalidArgument
	}

	return id, name, nil
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
