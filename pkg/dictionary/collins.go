package dictionary

import (
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

const (
	collinsURL      = "https://www.collinsdictionary.com/dictionary/english/%s"
	collinsSelector = "div.dictionary.dictentry div.definitions div.hom div.def"
)

type CollinsCrawler struct {
	Selector   string
	SearchURL  func(string) string
	SearchFunc func(results *[]string, counter *int) func(e *colly.HTMLElement)
	Logger     *log.Logger
}
