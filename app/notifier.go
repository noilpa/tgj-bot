package app

import (
	"fmt"
	"log"
	"time"

	ce "tgj-bot/customErrors"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	greeting = "Daily notification"
	cutoff   = "===================="
)

func (a *App) notify(timeout int64) {
	//
	// слать нотификации в определенное время Time
	// если не получены лайки и коменты за время Delay
	// пересчитывать лайки и комменты перед нотификацией
	//
	if a.Config.Notifier.IsAllow {
		go func() {
			var curDay time.Weekday
			var newDay time.Weekday
			var isNotified bool
			for t := range time.Tick(time.Minute) {
				newDay = t.Weekday()
				if newDay != curDay {
					curDay = newDay
					isNotified = false
				}

				// do not notify on weekends
				// in theory, the situation is impossible because the skipWeekends() method is used
				if newDay == time.Saturday || newDay == time.Sunday {
					isNotified = true
				}

				if approximatelyEqual(t, a.Config.Notifier) && !isNotified {
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
					for _, u := range us {
						mrStr, err := a.buildNotifierMRString(u.ID)
						if err != nil {
							log.Println(ce.Wrap(err, "notifier update reviews"))
							break
						}

						msg := fmt.Sprintf("@%s\n%s\n", u.TelegramUsername, cutoff)
						msg += mrStr

						log.Println(a.sendTgMessage(msg))

					}
					isNotified = true
				}
			}
		}()
	} else {
		log.Println("Notifications does not allow in config")
	}
}

func (a *App) sendTgMessage(msg string) (err error) {
	if m, err := a.Telegram.Bot.Send(tgbotapi.NewMessage(a.Config.Tg.ChatID, msg)); err != nil {
		log.Printf("Couldn't send message '%v': %v", m, err)
	}
	return
}

func approximatelyEqual(time time.Time, cfg NotifierConfig) bool {
	return time.Hour() == cfg.TimeHour && (time.Minute() > cfg.TimeMinute-1 && time.Minute() < cfg.TimeMinute+1)
}

func (a *App) buildNotifierMRString(uID int) (s string, err error) {
	rs, err := a.DB.GetOpenedReviewsByUserID(uID)
	if err != nil {
		err = ce.WrapWithLog(err, "notifier build message")
		return
	}

	for _, r := range rs {
		if time.Now().Unix() > r.UpdatedAt+a.Config.Notifier.Delay {
			mr, err := a.DB.GetMrByID(r.MrID)
			if err != nil {
				err = ce.WrapWithLog(err, "notifier build message")
				return
			}
			s += fmt.Sprintf("%s\n", mr.URL)
		}
	}
	return
}
