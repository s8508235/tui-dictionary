package dictionary

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/s8508235/tui-dictionary/pkg/cloudflare"
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

func (c *CollinsCrawler) Search(word string) ([]string, error) {

	result := make([]string, 0, 3)
	count := 0
	s, err := cloudflare.GetCloudFlareProtectedHTML(c.SearchURL(word))
	if err != nil {
		c.Logger.Errorln("not bypass cf", err)
		return result, err
	}
	// c.Logger.Info(s)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		c.Logger.Errorln("not get node", err)
		return result, err
	}
	doc.Find(c.Selector).Each(func(i int, s *goquery.Selection) {
		result = append(result, s.Text())
	})

	if count == 0 {
		return result, ErrorNoDef
	}
	return result, err
}
