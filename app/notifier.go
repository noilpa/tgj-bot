package app

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	ce "tgj-bot/custom_errors"
	"tgj-bot/external_service/jira"
	"tgj-bot/models"
)

const (
	greeting         = "🚀 Daily notification 🌞"
	cutoff           = "-----------------------"
	moveTaskToQAText = "please move task to QA:"
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
	//
	if !a.Config.Notifier.IsAllow {
		log.Println("Notifications does not allow in config")
		return
	}
	go func() {
		var curDay time.Weekday
		for t := range time.Tick(time.Duration(a.Config.Timings.CheckNotifyPeriod)) {
			lastSendNotify, err := a.loadLastSendNotify()
			if err != nil {
				a.logError(err)
				continue
			}

			curDay = t.Weekday()
			if lastSendNotify.Weekday() == curDay {
				continue
			}
			if curDay == time.Saturday || curDay == time.Sunday {
				continue
			}

			if t.Hour() >= a.Config.Notifier.TimeHour && t.Minute() >= a.Config.Notifier.TimeMinute {
				if err := a.sendDailyNotification(); err != nil {
					a.logError(err)
				}

				value := models.LastSendNotifyOption{Stamp: time.Now().Unix()}
				if err := a.DB.UpdateOptionByName(models.OptionLastSendNotify, value); err != nil {
					a.logError(err)
				}
			}
		}
	}()
}

func (a *App) loadLastSendNotify() (value time.Time, err error) {
	option, err := a.DB.LoadOptionByName(models.OptionLastSendNotify)
	if err != nil {
		return
	}

	var item models.LastSendNotifyOption
	err = json.Unmarshal([]byte(option.Item), &item)
	if err != nil {
		return
	}
	value = time.Unix(item.Stamp, 0)
	return
}

func (a *App) sendDailyNotification() error {
	us, err := a.DB.GetActiveUsers()
	if err != nil {
		log.Println(ce.Wrap(err, "notifier update reviews"))
		return err
	}
	log.Printf("Notifier active users %v", us)

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

	return nil
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

func (a *App) notifyReviewTask(mr models.MR) error {
	user, err := a.DB.GetUserByID(*mr.AuthorID)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("%s @%s, %s %s", readyToQAEmoji, user.TelegramUsername, moveTaskToQAText, mr.URL)
	if err := a.Telegram.SendMessage(msg); err != nil {
		return err
	}

	mr.NeedQANotify = false
	if _, err := a.DB.SaveMR(mr); err != nil {
		return err
	}

	return nil
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
