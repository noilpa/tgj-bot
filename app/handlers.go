package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tgj-bot/models"

	ce "tgj-bot/custom_errors"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var titleMRRegexp = regexp.MustCompile(`NC-\d+`)

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
	mrGitlabID, err := models.GetGitlabID(mrUrl)
	if err != nil {
		return
	}

	if a.isMrAlreadyExist(mrGitlabID) {
		return a.returnMrParty(mrGitlabID)
	}

	gitlabMR, err := a.Gitlab.GetMrByID(mrGitlabID)
	if err != nil {
		return
	}
	if !isMrTitleValid(gitlabMR.Title) {
		return errors.New("mr title must have ticket number in square brackets without spaces inside. Example:[NC-1234]")
	}
	author, err := a.DB.GetUserByGitlabID(gitlabMR.AuthorID)
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
		URL:            mrUrl,
		AuthorID:       &author.ID,
		GitlabID:       mrGitlabID,
		NeedJiraUpdate: true,
		NeedQANotify:   true,
	}
	mr, err = a.DB.CreateMR(mr)
	if err != nil {
		return
	}

	review := models.Review{
		MrID:      mr.ID,
		UpdatedAt: time.Now().Unix(),
	}

	msg := "New merge request " + randJoyEmoji() + "\n"
	reviewPartyBrief := make([]models.UserBrief, 0, len(reviewParty))
	for i := range reviewParty {

		// todo remove in next version
		if reviewParty[i].GitlabName == "" {
			reviewParty[i].GitlabName, err = a.Gitlab.GetUserByID(reviewParty[i].GitlabID)
			if err != nil {
				return ce.WrapWithLog(err, fmt.Sprintf("MR handler fail to get gitlab name for %v", reviewParty[i].TelegramUsername))
			}
			if _, err = a.DB.SaveUser(models.User{UserBrief: reviewParty[i].UserBrief}); err != nil {
				return ce.WrapWithLog(err, "MR handler fail")
			}
		}

		// todo remove in next version
		if reviewParty[i].GitlabID == 0 {
			reviewParty[i].GitlabID, err = a.Gitlab.GetUserByName(reviewParty[i].GitlabName)
			if err != nil {
				return ce.WrapWithLog(err, fmt.Sprintf("MR handler fail to get gitlab id for %v", reviewParty[i].TelegramUsername))
			}
			if _, err = a.DB.SaveUser(models.User{UserBrief: reviewParty[i].UserBrief}); err != nil {
				return ce.WrapWithLog(err, "MR handler fail")
			}
		}

		reviewPartyBrief = append(reviewPartyBrief, reviewParty[i].UserBrief)
		review.UserID = reviewParty[i].ID
		if err = a.DB.SaveReview(review); err != nil {
			return
		}
		msg += fmt.Sprintf("%s @%v\n", pointEmoji[i%2], reviewParty[i].TelegramUsername)
	}

	msg += cutoff + "\n" + mrUrl

	if err = a.Gitlab.WriteReviewers(mr.GitlabID, reviewPartyBrief); err != nil {
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
		return err
	}
	for _, mr := range mrs {
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

	closedMRs, err := a.DB.CloseMRs()
	if err != nil {
		return err
	}
	for _, mr := range closedMRs {
		if err = a.Gitlab.SetLabelToMR(mr.GitlabID, models.ReviewedLabel); err != nil {
			log.Printf("err set label for mr_id=%d: %v", mr.GitlabID, err)
			continue
		}
		log.Printf("successfully set label for mr_id=%d", mr.GitlabID)

		if mr.IsOnReview() && mr.NeedQANotify {
			if err := a.notifyReviewTask(mr); err != nil {
				log.Printf("err notify task mr_id=%d: %v", mr.GitlabID, err)
				continue
			}
		}
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
	userGitlabIDList, err := a.Gitlab.CheckMrComments(mr.GitlabID)
	if err != nil {
		return err
	}
	log.Printf("Check mr comments user's ids: %v", userGitlabIDList)
	for gitlabID, isCommented := range userGitlabIDList {
		u, err := a.DB.GetUserByGitlabID(gitlabID)
		if err != nil {
			ce.WrapWithLog(err, fmt.Sprintf("user not found by gitlab id=%d", gitlabID))
			continue
		}

		now := time.Now().Unix()
		now = a.skipWeekends(now)

		err = a.DB.UpdateReviewComment(models.Review{
			MrID:        mr.ID,
			UserID:      u.ID,
			IsCommented: isCommented,
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

		if err = a.DB.UpdateReview(models.Review{
			MrID:      mrID,
			UserID:    u.ID,
			UpdatedAt: time.Now().Unix(),
		}, user.ID); err != nil {
			log.Println(ce.Wrap(err, "Reallocate MRs UpdateReview"))
			continue
		}
		mr, err := a.DB.GetMrByID(mrID)
		if err != nil {
			log.Println(ce.Wrap(err, "Reallocate MRs GetMrByID"))
			continue
		}
		reviewers, err := a.DB.GetUsersByMrID(mrID)
		if err != nil {
			log.Println(ce.Wrap(err, "Reallocate MRs GetUsersByMrID"))
			continue
		}
		err = a.Gitlab.WriteReviewers(mr.GitlabID, reviewers)
		if err != nil {
			log.Println(ce.Wrap(err, "Reallocate MRs WriteReviewers"))
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

func (a *App) isMrAlreadyExist(mrID int) bool {
	mr, err := a.DB.GetMrByID(mrID)
	if err != nil {
		return false
	}
	defaultValue := models.MR{}
	if mr != defaultValue {
		return true
	}
	return false
}

func (a *App) returnMrParty(mrID int) (err error) {
	us, err := a.DB.GetUsersByMrID(mrID)
	msg := "Review party:\n"
	for i, u := range us {
		msg += fmt.Sprintf("%s %s\n", pointEmoji[i%2], u.TelegramUsername)
	}
	msg += cutoff + fmt.Sprintf("\n%s", a.createMrURL(mrID))
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

func (a *App) createMrURL(mrID int) string {
	return a.Config.Gl.MRBaseURL + "/" + strconv.Itoa(mrID)
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

func isMrTitleValid(title string) bool {
	indexes := titleMRRegexp.FindIndex([]byte(title))
	if len(indexes) == 0 {
		return true
	}
	if len(indexes) != 2 {
		return false
	}

	start := indexes[0]
	end := indexes[1]
	titleAmount := len(title)

	if start > 0 && string(title[start-1]) != "[" {
		return false
	}
	if end < (titleAmount-1) && string(title[end]) != "]" {
		return false
	}

	return true
}
