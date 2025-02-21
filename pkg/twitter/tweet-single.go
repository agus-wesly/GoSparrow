package twitter

import (
	"context"
	"encoding/json"
	"example/hello/pkg/core"
	"example/hello/pkg/terminal"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

type TweetSingleOption struct {
	Tweet
	TweetUrl string
	Context  context.Context
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

func (t *TweetSingleOption) handleSingleTweet() {
	err := chromedp.Run(t.Context,
		openTweetPage(t.TweetUrl),
	)
	if err != nil {
        panic(err)
	}
	t.scrollUntilBottom(t.Context)
}

func (t *TweetSingleOption) BeginSingleTweet() {
	defer func() {
		fmt.Println("Exporting to csv...")
		fileName := t.ExportToCSV()
		fmt.Println("Successfully exported at : ", fileName)
	}()

	ctx, acancel := core.CreateNewContext()
	defer acancel()
	t.Context = ctx

	err := chromedp.Run(t.Context,
		t.AttachAuthToken(),
	)
	if err != nil {
		log.Fatalln(err)
	}
	core.ListenEvent(ctx, "TweetDetail", func(byts []byte) {
		var tweetJson Response
		err = json.Unmarshal(byts, &tweetJson)
		if err == nil {
			fmt.Println("Got new tweet data 😎! Saving now ....")
			err = t.processTweetJSON(tweetJson)
		}
	},
	)
	t.handleSingleTweet()
}
