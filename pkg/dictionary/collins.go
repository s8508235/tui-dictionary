package dictionary

import (
	"fmt"
	"regexp"

	"github.com/s8508235/tui-dictionary/pkg/log"

	"github.com/gocolly/colly"
)

const URL = "https://www.collinsdictionary.com/dictionary/english/%s"

type Collins struct {
	Crawler *colly.Collector
	Logger  *log.Logger
}

var re = regexp.MustCompile(`(?s)[\s]+`)

func NewCollinsDictionary(logger *log.Logger) Dictionary {
	return &Collins{
		Crawler: colly.NewCollector(),
		Logger:  logger,
	}
}

func (c *Collins) Search(word string) ([3]string, int) {
	crawler := c.Crawler.Clone()
	var result [3]string
	count := 0
	crawler.OnHTML("div.dictionary.Cob_Adv_Brit.dictentry div.definitions div.hom div.def", func(e *colly.HTMLElement) {
		if count < 3 {
			result[count] = e.Text
		} else {
			return
		}
		count += 1
	})

	crawler.OnError(func(_ *colly.Response, err error) {
		c.Logger.Logrus.Debugln("Something went wrong:", err)
	})

	crawler.OnRequest(func(r *colly.Request) {
		c.Logger.Logrus.Debugln("Visiting", r.URL.String())
	})

	crawler.OnScraped(func(r *colly.Response) {
		c.Logger.Logrus.Debugln("Finished", r.Request.URL.String())
	})

	searchURL := fmt.Sprintf(URL, re.ReplaceAllString(word, "-"))
	crawler.Visit(searchURL)

	return result, count
}
