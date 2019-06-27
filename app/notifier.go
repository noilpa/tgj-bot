package app

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (a *App) notify(timeout int64) {
	//
	// раз в заданное время
	// пойти по всем МРам
	//    зайти в мр
	//    получить лайки и коменты
	//    выслать одно сообщение
	//

}

func (a *App) sendTgMessage(msg string) (err error) {
	if m, err := a.Telegram.Bot.Send(tgbotapi.NewMessage(a.Config.Tg.ChatID, msg)); err != nil {
		log.Printf("Couldn't send message '%v': %v", m, err)
	}
	return
}
