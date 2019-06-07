package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	es "tgj-bot/externalService"
)

type config struct {
	Tg es.TgConfig     `json:"telegram"`
	Gl es.GitlabConfig `json:"gitlab"`
	Jr es.JiraConfig   `json:"jira"`
}

func main() {
	cfg, err := readConfig("./conf/test_conf.json")
	if err != nil {
		panic(err)
	}

	err = es.RunBot(cfg.Tg)
	if err != nil {
		panic(err)
	}

}

func readConfig(path string) (config, error) {
	path, err := filepath.Abs(path)
	fmt.Println(path)
	if err != nil {
		return config{}, err
	}

	configFile, err := os.Open(path)
	if err != nil {
		return config{}, err
	}

	defer configFile.Close()

	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		return config{}, err
	}

	var cfg config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}
