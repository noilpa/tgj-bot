package app

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	ce "tgj-bot/custom_errors"
	"tgj-bot/external_service/jira"
	"tgj-bot/models"
)

const (
	greeting = "🚀 Daily notification 🌞"
	cutoff   = "-----------------------"
)

var (
	priorityEmoji = map[int]string{
		jira.PriorityHighest: "🔥",
		jira.PriorityHigh:    "🔥",
		jira.PriorityMedium:  "🔸",
		jira.PriorityLow:     "🔹",
		jira.PriorityLowest:  "🔹",
	}
	readyToQAEmoji = "✈️"
	pointEmoji     = []string{"🔹", "🔸"}
	sadEmoji       = []string{"😀", "😁", "😂", "🤣", "😃", "😄", "😅", "😆", "😉", "😊", "😋", "😎", "☺", "️🙂", "🤗", "🤔", "😐",
		"😑", "😶", "🙄", "😏", "😣", "😥", "😮", "🤐", "😯", "😪", "😫", "😴", "😌", "😛", "😜", "😝", "🤤", "😒", "😓",
		"😔", "😕", "🙃", "🤑", "😲", "☹", "️🙁", "😖", "😞", "😟", "😤", "😢", "😭", "😦", "😧", "😨", "😩"}
	joyEmoji = []string{"😉", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "☺", "️😊", "😇", "🙂", "🙃", "😉", "😌", "😍", "😘",
		"😗", "😙", "😚", "😋", "😜", "😜", "😝", "😛", "🤑", "🤗", "😎", "🤡", "🤠", "😏", "👐", "😸", "😹", "👻", "😺", "🙌",
		"👏", "🙏", "🤝", "👍", "👊", "✊", "🤞", "✌", "️🤘", "👈", "👉", "💪", "🤙", "👋", "🖖", "👑", "🌚", "🌝", "⭐", "️💫"}
)

func (a *App) notify() {
	//
	// слать нотификации в определенное время Time
	// если не получены лайки и коменты за время Delay
	// пересчитывать лайки и комменты перед нотификацией
	//
	if !a.Config.Notifier.IsAllow {
		log.Println("Notifications does not allow in config")
		return
	}
	go func() {
		var curDay time.Weekday
		var newDay time.Weekday
		var isNotified bool
		for t := range time.Tick(time.Minute) {
			newDay = t.Weekday()
			if newDay != curDay {
				curDay = newDay
				isNotified = false
				if curDay == time.Saturday || curDay == time.Sunday {
					isNotified = true
				}
			}

			if t.Hour() >= a.Config.Notifier.TimeHour && t.Minute() >= a.Config.Notifier.TimeMinute && !isNotified {
				// main logic
				if err := a.updateReviews(); err != nil {
					log.Println(ce.Wrap(err, "notifier update reviews"))
					continue
				}

				us, err := a.DB.GetActiveUsers()
				if err != nil {
					log.Println(ce.Wrap(err, "notifier update reviews"))
					continue
				}
				log.Printf("Notifier active users %v", us)

				mrs, err := a.DB.CloseMRs()
				if err != nil {
					log.Printf("err close mrs: %v", err)
				}
				for _, mr := range mrs {
					if err = a.Gitlab.SetLabelToMR(mr.GitlabID, models.ReviewedLabel); err != nil {
						log.Printf("err set label for mr_id=%d: %v", mr.GitlabID, err)
						continue
					}
					log.Printf("successfully set label for mr_id=%d", mr.GitlabID)
				}

				messagesCount := 0
				msg := greeting + "\n"
				for _, u := range us {
					qaTaskStr, err := a.buildNotifierQATask(u.ID)
					if err != nil {
						log.Println(ce.Wrap(err, "notifier qa task"))
						continue
					}

					mrStr, err := a.buildNotifierMRString(u.ID)
					if err != nil {
						log.Println(ce.Wrap(err, "notifier update reviews"))
						continue
					}

					log.Printf("Notifier string for user %d: %s %s\n", u.ID, mrStr, qaTaskStr)

					if len(mrStr) > 0 || len(qaTaskStr) > 0 {
						msg += fmt.Sprintf("%s\n@%s %s\n%s%s", cutoff, u.TelegramUsername, randSadEmoji(), qaTaskStr, mrStr)
						messagesCount++
					}
				}
				if messagesCount == 0 {
					msg += "\n" + a.praise()
				} else {
					msg += "\n" + a.motivate()
				}
				a.Telegram.SendMessage(msg)
				isNotified = true
			}
		}
	}()
}

func (a *App) buildNotifierQATask(uID int) (str string, err error) {
	mrs, err := a.DB.GetUserClosedMRs(uID, jira.StatusOnReview)
	if err != nil {
		err = ce.WrapWithLog(err, "notifier build message")
		return
	}
	log.Printf("User %d mrs:%d\n", uID, len(mrs))

	for _, mr := range mrs {
		str += fmt.Sprintf("%s %s\n", readyToQAEmoji, mr.URL)
	}

	return
}

func (a *App) buildNotifierMRString(uID int) (s string, err error) {
	rs, err := a.DB.GetOpenedReviewsByUserID(uID)
	if err != nil {
		err = ce.WrapWithLog(err, "notifier build message")
		return
	}
	log.Printf("User %d opened reviews: %v\n", uID, rs)

	for _, r := range rs {
		if time.Now().Unix() > r.UpdatedAt+a.Config.Notifier.Delay {
			mr, err := a.DB.GetMrByID(r.MrID)
			if err != nil {
				err = ce.WrapWithLog(err, "notifier build message")
				return s, err
			}
			s += fmt.Sprintf("%s %s\n", getPriorityEmoji(mr.JiraPriority), mr.URL)
		}
	}
	return
}

func randSadEmoji() string {
	return sadEmoji[rand.Intn(len(sadEmoji))]
}

func randJoyEmoji() string {
	return joyEmoji[rand.Intn(len(joyEmoji))]
}

func getPriorityEmoji(priority int) string {
	if icon, ok := priorityEmoji[priority]; ok {
		return icon
	}
	return ""
}

func (a App) praise() string {
	return a.Config.Notifier.Praise[rand.Intn(len(a.Config.Notifier.Praise))] + " " + randJoyEmoji()
}

func (a App) motivate() string {
	return a.Config.Notifier.Motivate[rand.Intn(len(a.Config.Notifier.Motivate))] + " " + randJoyEmoji()
}
