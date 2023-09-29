package dictionary

import (
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

const (
	cambridgeURL      = "https://dictionary.cambridge.org/dictionary/english/%s"
	cambridgeSelector = "div.entry div.sense-body div.def"
)

type cambridgeCrawler struct {
	Selector   string
	SearchURL  func(string) string
	SearchFunc func(results *[]string, counter *int) func(e *colly.HTMLElement)
	Logger     *log.Logger
}
