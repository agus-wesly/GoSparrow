package twitter

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/agus-wesly/GoSparrow/pkg/core"
	"github.com/agus-wesly/GoSparrow/pkg/terminal"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

var REACHING_LIMIT_ERR = errors.New("Reaching limit")

type TweetSingleOption struct {
	*Tweet
	TweetUrl string
	Context  *context.Context
}

func (t *TweetSingleOption) Prompt() {
	t.TweetUrl = "https://x.com/Taibandeng_/status/1890018993458881003"
	if !DEBUG {
		inp := terminal.Input{
			Message:   "Enter your desired twitter url",
			Validator: terminal.Required,
		}
		if err := inp.Ask(&t.TweetUrl); err != nil {
			panic(err)
		}
		inp = terminal.Input{
			Message:   "How many tweets do you want to retrieve ? [Default : 500]",
			Validator: terminal.IsNumber,
			Default:   "500",
		}
		if err := inp.Ask(&t.Limit); err != nil {
			panic(err)
		}
	}
}

func (t *TweetSingleOption) handleSingleTweet() error {
	err := chromedp.Run(*t.Context,
		t.openTweetPage(t.TweetUrl),
	)
	if err != nil {
		return err
	}
	t.Log.Success("Successfully Opened window")
	err = t.scrollUntilBottom(*t.Context)
	if err != nil {
		return err
	}
	return nil
}

func (t *TweetSingleOption) DemoLogging() {
	t.Log.Info("Opening page..........................")
	time.Sleep(2 * time.Second)
	t.Log.Success("Sucessfully opened page")

	for i := 0; i < 5; i++ {
		t.Log.Info("Scrolling down.....................")
		time.Sleep(1 * time.Second)
		if i == 3 {
			t.Log.Error("Something is wrong")
			break
		}
		t.Log.Success(fmt.Sprintf("Got new replies. Current Replies : %d", (i+1)*10))
	}
}

func (t *TweetSingleOption) BeginSingleTweet() {

	// t.DemoLogging()
	// if true {
	//     return
	// }

	defer func() {
		t.ExportToCSV()
	}()

	t.SetupToken()

	// TODO : use timeout context, idk maybe 3 minutes is enough ?
	ctx, acancel := core.CreateNewContext()
	defer acancel()

	err := chromedp.Run(ctx, t.AttachAuthToken())
	if err != nil {
		panic(err)
	}

	t.Context = &ctx

	core.ListenEvent(ctx, "TweetDetail", func(byts []byte) {
		var tweetJson Response
		err := json.Unmarshal(byts, &tweetJson)
		if err == nil {
			t.Log.Success("Got new tweet replies. Current replies get : ", len(t.TweetResults))
			err = t.processTweetJSON(tweetJson)
		}
	}, nil)
	t.handleSingleTweet()
	t.Log.Success("Finish scrapping. Total tweet received : ", len(t.TweetResults))
}

func (t *TweetSingleOption) scrollUntilBottom(ctx context.Context) error {
	var isAlreadyOnTheBottom bool = false
	for !isAlreadyOnTheBottom {
		if t.isReachingLimit() {
			return REACHING_LIMIT_ERR
		}
		err := chromedp.Run(ctx, t.scroll())
		if err != nil {
			break
		}
		chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			t.Log.Info("Scrolling down...")
			return nil
		}))
		chromedp.Run(ctx, chromedp.Evaluate(`Math.round(window.scrollY) + window.innerHeight >= document.body.scrollHeight`, &isAlreadyOnTheBottom))
	}
	return nil
}
