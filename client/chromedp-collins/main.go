package main

// https://www.zenrows.com/blog/web-scraping-golang#scraping-dynamic-content
import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	filename "github.com/keepeye/logrus-filename"
	"github.com/sirupsen/logrus"
)

type collinsDefinition struct {
	Def string
}

func main() {
	st := time.Now()
	// var collinsDefs []collinsDefinition
	logger := logrus.New()
	logger.Hooks.Add(filename.NewHook(logrus.InfoLevel))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx, cancel = signal.NotifyContext(ctx, syscall.SIGABRT, syscall.SIGTERM, os.Interrupt)
	defer cancel()
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("enable-automation", false),
	)
	ctx, cancel = chromedp.NewExecAllocator(
		ctx,
		opts...,
	)
	// defer cancel()
	// initializing a chrome instance
	pageCtx, cancel := chromedp.NewContext(
		// context.Background(),
		ctx,
		chromedp.WithLogf(logger.Infof),
		chromedp.WithDebugf(logger.Debugf),
		chromedp.WithErrorf(logger.Errorf),
	)
	defer cancel()

	// navigate to the target web page and select the HTML elements of interest
	var nodes []*cdp.Node
	// var res string
	if err := chromedp.Run(pageCtx,
		chromedp.Navigate("https://www.collinsdictionary.com/dictionary/english/choccy"),
		// chromedp.WaitVisible(`#challenge-form`, chromedp.ByID),
		// chromedp.EvaluateAsDevTools(`challengeForm.challengeResponse.value = generateToken()`, &res),
		// chromedp.Submit(`#challenge-form`, chromedp.ByID),
		// chromedp.Sleep(2*time.Second),
		// chromedp.Navigate("https://www.collinsdictionary.com/dictionary/english/test"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Nodes("div.content.definitions div.sense", &nodes, chromedp.ByQueryAll),
	); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		logger.Error(err)
		return
	}
	logger.Infoln("end navigate", time.Since(st).Seconds(), "secs", len(nodes))
	// ackCtx is created from pageCtx.
	// when ackCtx exceeds the deadline, pageCtx is not affected.
	ackCtx, cancel := context.WithTimeout(pageCtx, 20*time.Second)
	defer cancel()
	var def string
	defList := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if err := chromedp.Run(ackCtx,
			chromedp.Text("div.def", &def, chromedp.ByQuery, chromedp.FromNode(node)),
		); err != nil {
			logger.Error("ack error:", err)
		}
		defList = append(defList, def)
	}

	logger.Infoln(defList[0], len(defList))
}
