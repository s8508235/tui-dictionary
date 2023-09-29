package dictionary

import (
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

const (
	oxfordURL      = "https://www.oxfordlearnersdictionaries.com/definition/english/%s?q=%s"
	oxfordSelector = "div.entry li.sense span.def"
)

type oxfordCrawler struct {
	Selector   string
	SearchURL  func(string) string
	SearchFunc func(results *[]string, counter *int) func(e *colly.HTMLElement)
	Logger     *log.Logger
}
