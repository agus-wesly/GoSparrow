package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	ThreadedConversationsWithInjectionsV2 ThreadedConversation `json:"threaded_conversation_with_injections_v2"`
}

type ThreadedConversation struct {
	Instructions []Instruction `json:"instructions"`
}

type Instruction struct {
	Type    string  `json:"type"`
	Entries []Entry `json:"entries"`
}

type Entry struct {
	EntryId   string  `json:"entryId"`
	SortIndex string  `json:"sortIndex"`
	Content   Content `json:"content"`
}

type Content struct {
	EntryType   string       `json:"entryType"`
	ItemContent *ItemContent `json:"itemContent"`
	Items       *[]Items     `json:"items"`
}

type Items struct {
	EntryId string `json:"entryId"`
	Item    Item   `json:"item"`
}

type Item struct {
	ItemType     string       `json:"itemType"`
	TweetResults TweetResults `json:"tweet_results"`
	ItemContent  *ItemContent `json:"itemContent"`
}

type ItemContent struct {
	ItemType     string       `json:"itemType"`
	TweetResults TweetResults `json:"tweet_results"`
}

type TweetResults struct {
	Result Result `json:"result"`
}

type Result struct {
	RestId string `json:"rest_id"`
	Legacy Legacy `json:"legacy"`
	Core   Core   `json:"core"`
}

// Username //
type Core struct {
	Results UserResults `json:"user_results"`
}

type UserResults struct {
	Result UserResult `json:"result"`
}

type UserResult struct {
	Legacy UserResultLegacy `json:"legacy"`
}

type UserResultLegacy struct {
	Name string `json:"name"`
}

//

type Legacy struct {
	FullText  string `json:"full_text"`
	UserIdStr string `json:"user_id_str"`
}

type TweetScrapResult struct {
	TweetId   string `json:"tweet_id"`
	Author    string `json:"author_id"`
	Content   string `json:"content"`
	UserIdStr string `json:"user_id_str"`
}

func tweetListener(ctx context.Context, tweet_results map[string]TweetScrapResult) {
	// Listen phase
	var tweetRequestId network.RequestID = ""
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, "TweetDetail") {
                fmt.Println("received : ", responseReceivedEvent.RequestID)
				tweetRequestId = responseReceivedEvent.RequestID
			}
		case *network.EventLoadingFinished:
			if tweetRequestId == "" {
				break
			} else {
				tweetRequestId = ""
				fc := chromedp.FromContext(ctx)
				ctx2 := cdp.WithExecutor(ctx, fc.Target)
				var tweetJson Response
				go func() error {
					byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
					if err != nil {
						panic(err)
					}
					err = json.Unmarshal(byts, &tweetJson)
					if err != nil {
						return err
					}
                    fmt.Println("peeking : ", responseReceivedEvent.RequestID)
					fmt.Println("Got new tweet data 😎! Saving now ....")
					// saveToJsonFile(byts)
					processTweetJSON(tweetJson, tweet_results)
					return nil
				}()
			}
		}
	})
}

func tweetActions(ctx context.Context, url string) error {
	err := chromedp.Run(ctx,
		openPage(url),
	)
	if err != nil {
		return err
	}

	var isAlreadyOnTheBottom bool = false
	for !isAlreadyOnTheBottom {
		err := chromedp.Run(ctx, scrollDown())
		if err != nil {
			break
		}
		chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("Scrolling down...")
			return nil
		}))
		chromedp.Run(ctx, chromedp.Evaluate(`Math.round(window.scrollY) + window.innerHeight >= document.body.scrollHeight`, &isAlreadyOnTheBottom))
	}
	return nil
}

func handleSingleTweet(ctx context.Context) {
	err := chromedp.Run(ctx,
		authenticateTwitter("auth_token", tweetData.auth_token),
		verifyLoginTwitter(),
	)
	if err != nil {
		panic(err)
	}
	var tweet_results = make(map[string]TweetScrapResult)

	defer func() {
		fmt.Println("Exporting to csv...")
		fileName := exportTweetToCSV(tweet_results)
		fmt.Println("Successfully exported at : ", fileName)
	}()

	tweetListener(ctx, tweet_results)
	err = tweetActions(ctx, tweetData.tweet_url)
	if err != nil {
		fmt.Println("Error : ", err)
	}
}

func verifyLoginTwitter() chromedp.Tasks {
	var location string
	return chromedp.Tasks{
		chromedp.Navigate("https://x.com/explore"),
		chromedp.WaitReady(`body`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			act := chromedp.Location(&location)
			err := act.Do(ctx)
			if err != nil {
				return err
			}
			if strings.Contains(location, "login") {
				return errors.New("The auth token you provide is not valid")
			}
			return nil
		}),
	}

}

func authenticateTwitter(cookies ...string) chromedp.Tasks {
	if len(cookies)%2 != 0 {
		panic("Length must be divisible by 2")
	}
	expr := cdp.TimeSinceEpoch(time.Now().Add(3 * 24 * time.Hour))
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < len(cookies)-1; i++ {
				err := network.SetCookie(cookies[i], cookies[i+1]).
					WithDomain(".x.com").
					WithHTTPOnly(true).
					WithExpires(&expr).
					WithSecure(true).
					WithSameSite("strict").
					WithPath("/").
					Do(ctx)

				if err != nil {
					return err
				}
			}
			return nil
		}),
	}

}

func openPage(url string) chromedp.Tasks {
	fmt.Println("Opening window...")
	// Search for request that includes : TweetDetail
	// const url string = "https://x.com/jherr/status/1758571101964382487"
	tasks := chromedp.Tasks{
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitReady(`body [data-testid="tweetButtonInline"]`),
	}
	return tasks
}

var limitQueueLength int = 10

func processTweetJSON(jsonData Response, tweet_results map[string]TweetScrapResult) {
	var entries []Entry
	instructions := jsonData.Data.ThreadedConversationsWithInjectionsV2.Instructions
	if len(instructions) == 0 {
		return
	}
	entries = instructions[0].Entries
	for i := 0; i < len(entries); i++ {
		currentEntryContent := entries[i].Content
		var item *ItemContent
		if currentEntryContent.ItemContent != nil {
			item = currentEntryContent.ItemContent
			addToTweetResult(item, tweet_results)
		}

		if currentEntryContent.Items != nil {
			items := *currentEntryContent.Items
			for j := 0; j < len(items); j++ {
				item = items[j].Item.ItemContent
				addToTweetResult(item, tweet_results)
			}
		}
	}
}

func saveToJsonFile(data []byte) {
	err := os.WriteFile("tweet-response.json", data, 0644)
	if err != nil {
		log.Fatalln("error saving to json:", err)
	}
}

func addToTweetResult(item *ItemContent, tweet_results map[string]TweetScrapResult) TweetScrapResult {
	tweet := TweetScrapResult{
		TweetId:   item.TweetResults.Result.RestId,
		Author:    item.TweetResults.Result.Core.Results.Result.Legacy.Name,
		Content:   item.TweetResults.Result.Legacy.FullText,
		UserIdStr: item.TweetResults.Result.Legacy.UserIdStr,
	}
	tweet_results[tweet.TweetId] = tweet
	return tweet
}

// var tweet_results = make(map[string]TweetScrapResult)
// id,col1,col2
// id_1,340.384926,123.285031
// id_1,321.385028,4087.284675
func exportTweetToCSV(tweet_results map[string]TweetScrapResult) string {
	res := make([][]string, len(tweet_results)+1)
	res[0] = []string{"Tweet_Id", "Author", "Content"}

	var i int = 1
	for _, val := range tweet_results {
		res[i] = []string{val.TweetId, val.Author, val.Content}
		i++
	}
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.WriteAll(res)
	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}

	currentTime := time.Now().Local()
	fileName := fmt.Sprintf("res-%d.csv", currentTime.Unix())
	os.WriteFile(fileName, buf.Bytes(), 0644)
	return fileName
}
