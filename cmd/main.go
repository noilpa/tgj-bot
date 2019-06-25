package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	a "tgj-bot/app"
	"tgj-bot/externalService/database"
	gitlab_ "tgj-bot/externalService/gitlab"
	"tgj-bot/externalService/telegram"
)

func main() {
	var (
		app a.App
		err error
	)

	if len(os.Args) == 2 {
		app.Config, err = readConfig(os.Args[1])
	} else {
		app.Config, err = readConfig("../conf/test_conf.json")
	}
	if err != nil {
		log.Panic(err)
	}

	app.Telegram, err = telegram.RunBot(app.Config.Tg)
	if err != nil {
		log.Panic(err)
	}

	app.DB, err = database.RunDB(app.Config.Db)
	if err != nil {
		log.Panic(err)
	}
	defer app.DB.Close()

	app.Gitlab, err = gitlab_.RunGitlab(app.Config.Gl)

	err = app.Serve()
	log.Print("I'll be back...")
}

func readConfig(path string) (cfg a.Config, err error) {
	path, err = filepath.Abs(path)
	fmt.Println(path)
	if err != nil {
		return
	}
	configFile, err := os.Open(path)
	if err != nil {
		return
	}
	defer configFile.Close()
	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		return
	}
	if err = json.Unmarshal(data, &cfg); err != nil {
		return
	}
	return
}