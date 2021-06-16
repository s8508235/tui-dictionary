package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/s8508235/tui-dictionary/model"
	"github.com/s8508235/tui-dictionary/pkg/database"
	"github.com/s8508235/tui-dictionary/pkg/dictionary"
	"github.com/s8508235/tui-dictionary/pkg/log"
	"github.com/tcnksm/go-input"
	"github.com/vmihailenco/msgpack"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"
)

var errWord = errors.New("should be a word")

func main() {
	logger := log.New()
	cfg, err := ini.Load("app.ini")
	if err != nil {
		logger.Logrus.Errorln("Fail to read file", err)
		os.Exit(1)
	}
	dictionaryType := cfg.Section("").Key("dictionary").String()
	target := strings.ToLower(strings.ReplaceAll(cfg.Section("").Key("target").String(), " ", "-"))

	logLevel := cfg.Section("").Key("level").String()
	logger.SetLogLevel(logLevel)

	if len(target) == 0 {
		target = "target"
	}
	fmt.Println("Use dictionary:", dictionaryType)
	fmt.Println("Target:", target)
	fmt.Println("===== Press Ctrl+C to exit =====")
	db, err := database.NewSqlLiteConnection(target, logger)
	if err != nil {
		logger.Logrus.Errorln("Fail to open db", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&model.Dictionary{})
	if err != nil {
		logger.Logrus.Errorln("Fail to migrate db", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Logrus.Errorln("Fail to access sql/db", err)
		os.Exit(1)
	}
	if err := sqlDB.Ping(); err != nil {
		logger.Logrus.Errorln("Fail to ping db", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		logger.Logrus.Errorln("Fail to init lemmatizer", err)
		os.Exit(1)
	}
	dict, err := dictionary.NewMyPreferDictionary(logger, "tcp", "dict.dict.org:2628", "!")
	if err != nil {
		logger.Logrus.Errorln("Fail to init dictionary", err)
		os.Exit(1)
	}

	query := "word"

	for {
		inputWord, err := ui.Ask(query, &input.Options{
			Required:    true,
			Loop:        true,
			HideDefault: true,
			HideOrder:   true,
			ValidateFunc: func(s string) error {
				match, err := regexp.MatchString(`(?s)^[a-zA-Z\s]+$`, s)
				if err != nil {
					logger.Logrus.Errorln("Fail to match pattern", err)
					return err
				}
				if match {
					return nil
				} else {
					logger.Logrus.Errorln(errWord)
					return errWord
				}
			},
		})
		switch err {
		case nil:
			break
		case errWord:
			continue
		case input.ErrInterrupted:
			fmt.Println()
			os.Exit(0)
		default:
			logger.Logrus.Errorln("Fail to ask", err)
			os.Exit(1)
		}
		searchWord := lemmatizer.Lemma(inputWord)
		logger.Logrus.Debugln("going to search", searchWord)
		var word model.Dictionary
		result := db.Where("word = ?", searchWord).First(&word)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			results, err := dict.Search(searchWord)
			if err == dictionary.ErrorNoDef {
				fmt.Printf("no definition for: %s\n", searchWord)
				continue
			} else if err != nil {
				logger.Logrus.Errorln("Search error:", err)
				os.Exit(1)
			}
			defs, err := displayDefinition(results...)
			if err != nil {
				logger.Logrus.Error(err)
				continue
			}
			fmt.Printf("definition: %s\n", defs)
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
			defs, err := displayDefinition(results...)
			if err != nil {
				logger.Logrus.Error(err)
				continue
			}
			fmt.Printf("definition: %s\n", defs)
		}
	}
}

func displayDefinition(defs ...string) (string, error) {
	var buf strings.Builder
	var err error
	_, err = buf.WriteRune('\n')
	if err != nil {
		return buf.String(), err
	}
	for i, def := range defs {
		if len(def) > 0 {
			_, err = buf.WriteString(strconv.Itoa(i + 1))
			if err != nil {
				break
			}
			_, err = buf.WriteString(". ")
			if err != nil {
				break
			}
			_, err = buf.WriteString("\t")
			if err != nil {
				break
			}
			_, err = buf.WriteString(strings.ReplaceAll(def, "\n", "\n\t"))
			if err != nil {
				break
			}
			_, err = buf.WriteRune('\n')
			if err != nil {
				break
			}
		} else {
			break
		}
	}
	return buf.String(), err
}
