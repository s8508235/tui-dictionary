package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c-bata/go-prompt"
	"github.com/s8508235/tui-dictionary/model"
	"github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/s8508235/tui-dictionary/pkg/tools"

	"github.com/s8508235/tui-dictionary/pkg/database"
	"gopkg.in/ini.v1"
)

func main() {
	defer tools.Exit()
	logger := log.New()

	cfg, err := ini.Load("app.ini")
	if err != nil {
		logger.Logrus.Errorln("Fail to read file:", err)
		return
	}
	targetsSection := cfg.Section("").Key("targets")
	targets := targetsSection.Strings(",")
	p := prompt.New(func(t string) {},
		targetsCompleter(targets),
		prompt.OptionPrefix("target: "))
	target := p.Input()

	logLevel := cfg.Section("").Key("level").String()
	if logLevel == "" {
		logLevel = "info"
	}
	if len(target) == 0 {
		target = "target"
	}
	db, err := database.NewSqlLiteConnection(target, logger)
	if err != nil {
		logger.Logrus.Errorln("Fail to open db:", err)
		return
	}

	err = db.AutoMigrate(&model.Dictionary{})
	if err != nil {
		logger.Logrus.Errorln("Fail to migrate db:", err)
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Logrus.Errorln("Fail to access sql/db:", err)
		return
	}
	if err := sqlDB.Ping(); err != nil {
		logger.Logrus.Errorln("Fail to ping db:", err)
		return
	}
	defer sqlDB.Close()

	file, err := os.OpenFile(filepath.Clean(fmt.Sprintf("%s.txt", target)), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		logger.Logrus.Errorln("Fail to open output file", err)
		return
	}
	defer file.Close()

	dictionaries := []model.Dictionary{}
	result := db.Find(&dictionaries)
	if result.Error != nil {
		logger.Logrus.Errorln("Fail to full scan db:", err)
		return
	}

	logger.Logrus.Infoln("start migrate", target, "to text file", "with", result.RowsAffected, "word(s)")
	for _, dictionary := range dictionaries {
		if _, err := file.WriteString(dictionary.Word + "\n"); err != nil {
			logger.Logrus.Errorln("Fail to write output file", err)
			return
		}
	}
	logger.Logrus.Infoln("end migrate")
}

func targetsCompleter(targets []string) func(d prompt.Document) []prompt.Suggest {
	return func(d prompt.Document) []prompt.Suggest {
		s := make([]prompt.Suggest, 0, len(targets))
		for _, target := range targets {
			s = append(s, prompt.Suggest{Text: target})
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}
}
