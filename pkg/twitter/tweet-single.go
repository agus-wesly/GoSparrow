package twitter

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/agus-wesly/GoSparrow/pkg/core"
	"github.com/agus-wesly/GoSparrow/pkg/terminal"

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

func (t *TweetSingleOption) BeginSingleTweet() {

	defer func() {
		t.ExportToCSV()
	}()

	if !t.ValidateTweetUrl(t.TweetUrl) {
		t.Log.Error("Invalid url. Please provide valid tweet url.")
		os.Exit(1)
	}

	t.SetupToken()
	ctx, acancel := core.CreateNewContextWithTimeout(3 * time.Minute)
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
	err = t.handleSingleTweet()
	if err != nil {
		t.Log.Error(err.Error())
		os.Exit(1)
	} else {
		t.Log.Success("Finish scrapping. Total tweet received : ", len(t.TweetResults))
	}
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
