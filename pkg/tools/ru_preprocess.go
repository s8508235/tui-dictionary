package tools

import (
	"fmt"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

// preprocess -> strip accent -> write back
// RussianPreprocess relies on https://russiangram.com/ to get vocabulary with the accent mark, pick always first one
// despite changing the accents can not only complicate the communication, but change the meaning of a word completely.
func RussianPreprocess(word string) (string, error) {
	c := colly.NewCollector()
	extensions.RandomUserAgent(c)
	var viewState string
	var viewStateGenerator string
	var eventTarget string
	var eventTargetArgument string
	var eventValidation string
	var result string
	c.OnRequest(func(r *colly.Request) {
		// fmt.Println("Visiting", r.URL)
	})
	c.OnHTML("div.aspNetHidden input#__VIEWSTATE", func(e *colly.HTMLElement) {
		viewState = e.Attr("value")
		// fmt.Println("view state:", viewState)
	})
	c.OnHTML("div.aspNetHidden input#__VIEWSTATEGENERATOR", func(e *colly.HTMLElement) {
		viewStateGenerator = e.Attr("value")
		// fmt.Println("event target:", viewStateGenerator)
	})
	c.OnHTML("div.aspNetHidden input#__EVENTTARGET", func(e *colly.HTMLElement) {
		eventTarget = e.Attr("value")
		// fmt.Println("event target:", eventTarget)
	})
	c.OnHTML("div.aspNetHidden input#__EVENTARGUMENT", func(e *colly.HTMLElement) {
		eventTargetArgument = e.Attr("value")
		// fmt.Println("view state:", eventTargetArgument)
	})
	c.OnHTML("div.aspNetHidden input#__EVENTVALIDATION", func(e *colly.HTMLElement) {
		eventValidation = e.Attr("value")
		// fmt.Println("view state:", eventValidation)
	})
	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})
	c.Visit("https://russiangram.com/")

	c.OnHTML("textarea.input-textbox", func(e *colly.HTMLElement) {
		result = e.Text
	})
	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})
	var formData = map[string]string{
		`__VIEWSTATE`:                           viewState,
		`__VIEWSTATEGENERATOR`:                  viewStateGenerator,
		`__EVENTTARGET`:                         eventTarget,
		`__EVENTARGUMENT`:                       eventTargetArgument,
		`__EVENTVALIDATION`:                     eventValidation,
		"ctl00$MainContent$UserSentenceTextbox": word,
		"ctl00$MainContent$SubmitButton":        "Annotate",
	}
	err := c.Post("https://russiangram.com/", formData)

	return result, err
}
