package twitter

import (
	"context"
	"encoding/json"
	"example/hello/pkg/core"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type TweetSingleOption struct {
    Tweet
	TweetUrl string
    Context context.Context
}

func (t *TweetSingleOption) Prompt() {
	t.TweetUrl = "https://x.com/Taibandeng_/status/1890018993458881003"
	if !DEBUG {
		fmt.Print("Enter your desired twitter url : ")
		fmt.Scanln(&t.TweetUrl)
		fmt.Print("How many tweets do you want to retrieve ? [Default : 500] : ")
		fmt.Scanln(&t.Limit)
	}
}


func (t *TweetSingleOption) handleSingleTweet() {
	err := chromedp.Run(t.Context,
		openTweetPage(t.TweetUrl),
	)
	if err != nil {
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
	t.Listener(ctx, "TweetDetail", func(requestId network.RequestID, ctx2 context.Context) {
		go func() error {
			byts, err := network.GetResponseBody(requestId).Do(ctx2)
			if err != nil {
				fmt.Println("No resource error")
				return err
			}
			var tweetJson Response
			err = json.Unmarshal(byts, &tweetJson)
			if err == nil {
				fmt.Println("Got new tweet data 😎! Saving now ....")
				err = t.processTweetJSON(tweetJson)
			}
			return nil
		}()
	},
	)
	t.handleSingleTweet()
}
