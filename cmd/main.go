package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"tgj-bot/app"
	"tgj-bot/externalService/database"
	"tgj-bot/externalService/telegram"
)

func main() {
	var (
		appCtx app.App
		err    error
	)

	if len(os.Args) == 2 {
		appCtx.Config, err = readConfig(os.Args[1])
	} else {
		appCtx.Config, err = readConfig("../conf/test_conf.json")
	}
	if err != nil {
		log.Panic(err)
	}
	appCtx.Telegram, err = telegram.RunBot(appCtx.Config.Tg)
	if err != nil {
		log.Panic(err)
	}
	appCtx.DB, err = database.RunDB(appCtx.Config.Db)
	if err != nil {
		log.Panic(err)
	}
	defer appCtx.DB.Close()
	err = appCtx.Serve()
	log.Print("I'll be back...")
}

func readConfig(path string) (cfg app.Config, err error) {
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
