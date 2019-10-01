package app

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	ce "tgj-bot/custom_errors"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	greeting = "🚀 Daily notification 🌞"
	cutoff   = "-----------------------"
)

var (
	point = []string{"🔹", "🔸"}
	emoji = []string{"😀", "😁", "😂", "🤣", "😃", "😄", "😅", "😆", "😉", "😊", "😋", "😎", "☺", "️🙂", "🤗", "🤔", "😐",
		"😑", "😶", "🙄", "😏", "😣", "😥", "😮", "🤐", "😯", "😪", "😫", "😴", "😌", "😛", "😜", "😝", "🤤", "😒", "😓",
		"😔", "😕", "🙃", "🤑", "😲", "☹", "️🙁", "😖", "😞", "😟", "😤", "😢", "😭", "😦", "😧", "😨", "😩"}
)

func (a *App) notify() {
	//
	// слать нотификации в определенное время Time
	// если не получены лайки и коменты за время Delay
	// пересчитывать лайки и комменты перед нотификацией
	//
	if !a.Config.Notifier.IsAllow {
		log.Println("Notifications does not allow in config")
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
					break
				}

				us, err := a.DB.GetActiveUsers()
				if err != nil {
					log.Println(ce.Wrap(err, "notifier update reviews"))
					break
				}
				log.Println(a.sendTgMessage(greeting))
				log.Printf("Notifier active users %v", us)

				for _, u := range us {
					mrStr, err := a.buildNotifierMRString(u.ID)
					if err != nil {
						log.Println(ce.Wrap(err, "notifier update reviews"))
						break
					}
					log.Printf("Notifier MR string for user %d: %v\n", u.ID, mrStr)

					if mrStr != "" {
						msg := fmt.Sprintf("@%s %s\n%s\n", u.TelegramUsername, randEmoji(), cutoff)
						msg += mrStr

						log.Println(a.sendTgMessage(msg))
					}
				}
				isNotified = true
			}
		}
	}()

}

func (a *App) sendTgMessage(msg string) (err error) {
	if m, err := a.Telegram.Bot.Send(tgbotapi.NewMessage(a.Config.Tg.ChatID, msg)); err != nil {
		log.Printf("Couldn't send message '%v': %v", m, err)
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

	for i, r := range rs {
		if time.Now().Unix() > r.UpdatedAt+a.Config.Notifier.Delay {
			mr, err := a.DB.GetMrByID(r.MrID)
			if err != nil {
				err = ce.WrapWithLog(err, "notifier build message")
				return s, err
			}
			s += fmt.Sprintf("%s %s\n", string(point[i%2]), mr.URL)
		}
	}
	return
}

func randEmoji() string {
	return emoji[rand.Intn(len(emoji))]
}
