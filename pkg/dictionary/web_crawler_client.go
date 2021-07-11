package dictionary

import (
	"github.com/s8508235/tui-dictionary/pkg/log"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

// const webDictionaryCrawlerURL = "https://www.collinsdictionary.com/dictionary/english/%s"

type WebDictionaryCrawler struct {
	Selector   string
	SearchURL  func(string) string
	SearchFunc func(results *[]string, counter *int) func(e *colly.HTMLElement)
	Crawler    *colly.Collector
	Logger     *log.Logger
}

func (c *WebDictionaryCrawler) Search(word string) ([]string, error) {
	crawler := c.Crawler.Clone()
	// https://github.com/gocolly/colly/issues/150
	extensions.RandomUserAgent(crawler)
	result := make([]string, 0, 3)
	count := 0
	crawler.OnHTML(c.Selector, c.SearchFunc(&result, &count))

	crawler.OnError(func(_ *colly.Response, err error) {
		c.Logger.Logrus.Debugln("Something went wrong:", err)
	})

	crawler.OnRequest(func(r *colly.Request) {
		c.Logger.Logrus.Debugln("Visiting", r.URL.String())
	})

	crawler.OnScraped(func(r *colly.Response) {
		c.Logger.Logrus.Debugln("Finished", r.Request.URL.String())
	})

	err := crawler.Visit(c.SearchURL(word))

	if count == 0 {
		return result, ErrorNoDef
	}
	return result, err
}
