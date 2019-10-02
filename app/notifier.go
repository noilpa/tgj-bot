package app

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	ce "tgj-bot/custom_errors"
)

const (
	greeting = "🚀 Daily notification 🌞"
	cutoff   = "-----------------------"
)

var (
	pointEmoji = []string{"🔹", "🔸"}
	sadEmoji   = []string{"😀", "😁", "😂", "🤣", "😃", "😄", "😅", "😆", "😉", "😊", "😋", "😎", "☺", "️🙂", "🤗", "🤔", "😐",
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

				messagesCount := 0
				msg := greeting + "\n"
				for _, u := range us {
					mrStr, err := a.buildNotifierMRString(u.ID)
					if err != nil {
						log.Println(ce.Wrap(err, "notifier update reviews"))
						continue
					}
					log.Printf("Notifier MR string for user %d: %v\n", u.ID, mrStr)

					if mrStr != "" {
						msg += fmt.Sprintf("%s\n@%s %s\n%s", cutoff, u.TelegramUsername, randSadEmoji(), mrStr)
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

func (a *App) buildNotifierMRString(uID int) (s string, err error) {
	rs, err := a.DB.GetOpenedReviewsByUserID(uID)
	if err != nil {
		err = ce.WrapWithLog(err, "notifier build message")
		return
	}
	log.Printf("User %d opened reviews: %v\n", uID, rs)

	for i, r := range rs {
		if time.Now().Unix() > r.UpdatedAt+a.Config.Notifier.Delay {
			mr, err := a.DB.GetMrByID(r.MrID)
			if err != nil {
				err = ce.WrapWithLog(err, "notifier build message")
				return s, err
			}
			s += fmt.Sprintf("%s %s\n", pointEmoji[i%2], mr.URL)
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

func (a App) praise() string {
	return a.Config.Notifier.Praise[rand.Intn(len(a.Config.Notifier.Praise))] + randJoyEmoji()
}

func (a App) motivate() string {
	return a.Config.Notifier.Motivate[rand.Intn(len(a.Config.Notifier.Motivate))] + randJoyEmoji()
}
