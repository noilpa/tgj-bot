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
					msg += "\n" + praise()
				} else {
					msg += "\n" + motivate()
				}
				log.Println(a.sendTgMessage(msg))
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

func praise() string {
	templates := []string{
		"Мне так приятно заходить в такой отревьюиный проект)",
		"Я вижу, что вы очень постарались!",
		"Спасибо за ревью и дай бог здоровья!",
		"Спасибо большое за вашу помощь в ревью!",
		"Хорошая работа, так держать!",
	}
	return templates[rand.Intn(len(templates))] + randJoyEmoji()
}

func motivate() string {
	templates := []string{
		"Just Do IT",
		"«Не ноша тянет вас вниз, а то, как вы ее несете», — Лу Хольц",
		"«Я не боюсь умереть, но я боюсь не попытаться», — Jay Z",
		"«Притворяйся, пока не получится! Делай вид, что ты настолько уверен в себе, насколько это необходимо, пока не обнаружишь, что так оно и есть», — Брайан Трейси",
		"«Всегда выкладывайся на полную. Что посеешь — то и пожнешь», — Ог Мандино",
	}
	return templates[rand.Intn(len(templates))] + randJoyEmoji()
}
