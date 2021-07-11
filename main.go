package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/briandowns/spinner"
	"github.com/c-bata/go-prompt"
	"github.com/s8508235/tui-dictionary/model"
	"github.com/s8508235/tui-dictionary/pkg/database"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/s8508235/tui-dictionary/pkg/tools"
	"github.com/vmihailenco/msgpack"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"
)

func main() {
	defer tools.Exit()
	logger := log.New()
	cfg, err := ini.Load("app.ini")
	if err != nil {
		logger.Logrus.Errorln("Fail to read file:", err)
		return
	}
	target := strings.ToLower(strings.ReplaceAll(cfg.Section("").Key("target").String(), " ", "-"))

	logLevel := cfg.Section("").Key("level").String()
	logger.SetLogLevel(logLevel)

	if len(target) == 0 {
		target = "target"
	}
	starter := func() {
		fmt.Println("Target:", target)
		fmt.Println("===== Input q to exit =====")
		fmt.Println("===== Input t to toggle hidden mode =====")
	}
	starter()
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

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		logger.Logrus.Errorln("Fail to init lemmatizer:", err)
		return
	}
	dict, err := dictionary.NewMyPreferDictionary(logger)
	if err != nil {
		logger.Logrus.Errorln("Fail to init dictionary:", err)
		return
	}

	query := "Word: "
	var buf bytes.Buffer
	var flag bool
	s := spinner.New(spinner.CharSets[43], 100*time.Millisecond) // Build our new spinner
	s.FinalMSG = "\r"
	for {
		s.Stop()
		inputWord := prompt.Input(query, completer)
		err := tools.WordValidate(inputWord)
		switch err {
		case nil:
		case tools.ErrWord:
			logger.Logrus.Errorln("Wrong format:", err)
			continue
		default:
			logger.Logrus.Errorln("Fail to ask:", err)
			return
		}
		inputWord = strings.TrimSuffix(inputWord, "\r")
		switch inputWord {
		case "cls":
			if err := tools.Clear(); err != nil {
				logger.Logrus.Errorln("Fail to clear screen:", err)
				continue
			}
			starter()
			continue
		case "q", "Q":
			fmt.Println()
			return
		case "t", "T":
			flag = !flag
			if !flag {
				fmt.Print(buf.String())
				buf.Reset()
			}
			continue
		}
		s.Start()
		cols, err := tools.Cols()
		if err != nil {
			logger.Logrus.Errorln("Fail to get terminal info:", err)
			return
		}
		lines, err := tools.Lines()
		if err != nil {
			logger.Logrus.Errorln("Fail to get terminal info:", err)
			return
		}
		searchWord := lemmatizer.Lemma(inputWord)
		logger.Logrus.Debugln("going to search", searchWord)
		var word model.Dictionary
		dbResult := db.Where("word = ?", searchWord).First(&word)
		if errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
			results, err := dict.Search(searchWord)
			if err == dictionary.ErrorNoDef {
				fmt.Printf("no definition for: %s\n", searchWord)
				continue
			} else if err != nil {
				logger.Logrus.Errorln("Search error:", err)
				return
			}
			defs, err := tools.DisplayDefinition(logger, lines, cols, results...)
			if err != nil {
				logger.Logrus.Error(err)
				continue
			}
			s.Stop()
			if flag {
				fmt.Fprintf(&buf, "definition of %s: %s\n", inputWord, defs)

			} else {
				fmt.Printf("definition of %s: %s\n", inputWord, defs)
			}
			b, err := msgpack.Marshal(results)
			if err != nil {
				logger.Logrus.Error(err)
				continue
			}
			db.Create(&model.Dictionary{Word: searchWord, Definition: b})
		} else {
			var results []string
			err = msgpack.Unmarshal(word.Definition, &results)
			if err != nil {
				logger.Logrus.Error(err)
				continue
			}
			defs, err := tools.DisplayDefinition(logger, lines, cols, results...)
			if err != nil {
				logger.Logrus.Error(err)
				continue
			}
			s.Stop()
			if flag {
				fmt.Fprintf(&buf, "definition of %s: %s\n", inputWord, defs)
			} else {
				fmt.Printf("definition of %s: %s\n", inputWord, defs)
			}
		}
	}

}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
