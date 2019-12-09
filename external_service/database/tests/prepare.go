package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"tgj-bot/external_service/database"
)

type TestConfig struct {
	Database     database.DbConfig `json:"database"`
	TemplateConf string            `json:"tmpl_conf"`
	DestConf     string            `json:"dest_conf"`
}

type Config struct {
	Tests TestConfig `json:"tests"`
}

func main() {
	conf, err := readConfig(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if err := createConf(conf); err != nil {
		log.Fatal(err)
	}
}

func createConf(conf *Config) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	tmpl, err := os.Open(dir + "/" + conf.Tests.TemplateConf)
	if err != nil {
		return err
	}
	defer tmpl.Close()

	contentBytes, err := ioutil.ReadAll(tmpl)
	if err != nil {
		return err
	}

	content := string(contentBytes)

	replaceData := map[string]string{
		"PG_DRIVER_NAME":   conf.Tests.Database.DriverName,
		"PG_HOST":          conf.Tests.Database.Host,
		"PG_PORT":          conf.Tests.Database.Port,
		"PG_USER":          conf.Tests.Database.User,
		"PG_PASS":          conf.Tests.Database.Pass,
		"PG_DBNAME":        conf.Tests.Database.DBName,
		"PG_MIGRATION_DIR": dir + "/" + conf.Tests.Database.MigrationsDir,
	}

	for key, data := range replaceData {
		content = strings.Replace(content, "%("+key+")%", data, -1)
	}
	if err := ioutil.WriteFile(conf.Tests.DestConf, []byte(content), 0644); err != nil {
		return err
	}

	return nil
}

func readConfig(path string) (cfg *Config, err error) {
	path, err = filepath.Abs(path)
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
