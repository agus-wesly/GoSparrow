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
	"slices"
	"strings"
	"sync"
	"time"

	"example/hello/pkg/twitter"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func tweetListener(ctx context.Context, tweet_results map[string]twitter.TweetScrapResult) {
	tweetRequestIdList := make([]network.RequestID, 0)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, "TweetDetail") {
                tweetRequestIdList = append(tweetRequestIdList, responseReceivedEvent.RequestID)
			}
		case *network.EventLoadingFinished:
			if !slices.Contains(tweetRequestIdList, responseReceivedEvent.RequestID) {
				break
			} else {
                tweetRequestIdList = slices.DeleteFunc(tweetRequestIdList, func(targetId network.RequestID) bool {
                    return targetId == responseReceivedEvent.RequestID
                })
				fc := chromedp.FromContext(ctx)
				ctx2 := cdp.WithExecutor(ctx, fc.Target)
				var tweetJson twitter.Response
				go func() error {
					byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
					if err != nil {
						fmt.Println("No resource error")
                        panic(err)
					}
					err = json.Unmarshal(byts, &tweetJson)
					if err == nil {
						fmt.Println("Got new tweet data 😎! Saving now ....")
						processTweetJSON(tweetJson, tweet_results)
					}
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

// Todo : Make this function only : accept url and then return replies
func handleSingleTweet(ctx context.Context, tweetUrl string) {
	err := tweetActions(ctx, tweetUrl)
	if err != nil {
		fmt.Println("Error : ", err)
	}
}

func beginSingleTweet(tweetUrl string) {
	var tweet_results = make(map[string]twitter.TweetScrapResult)

	defer func() {
		fmt.Println("Exporting to csv...")
		fileName := exportTweetToCSV(tweet_results)
		fmt.Println("Successfully exported at : ", fileName)
	}()

	ctx, acancel := createNewContext()
	defer acancel()
	err := chromedp.Run(ctx,
		authenticateTwitter("auth_token", tweetData.AuthToken),
	)
	if err != nil {
		log.Fatalln(err)
	}
	tweetListener(ctx, tweet_results)
	handleSingleTweet(ctx, tweetUrl)
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

func processTweetJSON(jsonData twitter.Response, tweet_results map[string]twitter.TweetScrapResult) {
	var entries []twitter.Entry
	instructions := jsonData.Data.ThreadedConversationsWithInjectionsV2.Instructions
	if len(instructions) == 0 {
		return
	}
	entries = instructions[0].Entries
	for i := 0; i < len(entries); i++ {
		currentEntryContent := entries[i].Content
		var item *twitter.ItemContent
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

// Todo : do again until cleaned all
func cleanupContent(content string) string {
	if content == "" || content[0] != '@' {
		return content
	}
	idx := 0
	for idx < len(content) && content[idx] != ' ' {
		idx++
	}
	res := strings.Trim(content[idx:], " ")
	return res
}

func addToTweetResult(item *twitter.ItemContent, tweet_results map[string]twitter.TweetScrapResult) twitter.TweetScrapResult {
	tweet := twitter.TweetScrapResult{
		TweetId:   item.TweetResults.Result.RestId,
		Author:    item.TweetResults.Result.Core.Results.Result.Legacy.Name,
		Content:   cleanupContent(item.TweetResults.Result.Legacy.FullText),
		UserIdStr: item.TweetResults.Result.Legacy.UserIdStr,
	}
	tweet_results[tweet.TweetId] = tweet
	return tweet
}

func exportTweetToCSV(tweet_results map[string]twitter.TweetScrapResult) string {
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

func scrollDown() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight})`, nil),
		chromedp.Evaluate(`document.evaluate("//span[contains(., 'Show replies')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext()?.click()`, nil),
		chromedp.Evaluate(`document.evaluate("//button[contains(., 'Show replies')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext()?.click()`, nil),
		chromedp.Sleep(2 * time.Second),
	}
}

func beginSearchTweet(opt *twitter.SearchOption) {
	ctx, cancel := createNewContext()
	// ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	searchUrl := constructSearchUrl(opt)
	fmt.Println(searchUrl)
	var tweet_results = make(map[string]twitter.TweetScrapResult)
	var single_tweets = make([]string, 0)
	tweetSearchListener(ctx, &single_tweets, tweet_results, &wg)
	err := chromedp.Run(ctx,
		authenticateTwitter("auth_token", tweetData.AuthToken),
		network.Enable(),
		chromedp.Navigate(searchUrl),
		chromedp.WaitReady(`body [aria-label='Timeline: Search timeline']`),
	)
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		fmt.Println("Exporting to csv...")
		fileName := exportTweetToCSV(tweet_results)
		fmt.Println("Successfully exported at : ", fileName)
	}()

	wg.Wait()
	tweetListener(ctx, tweet_results)
	for i, tweetUrl := range single_tweets {
		fmt.Println("Current idx : ", i)
		handleSingleTweet(ctx, tweetUrl)
	}
}

func tweetSearchListener(ctx context.Context, single_tweets *[]string, tweet_results map[string]twitter.TweetScrapResult, wg *sync.WaitGroup) {
	var tweetRequestId network.RequestID = ""
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, "SearchTimeline") {
				tweetRequestId = responseReceivedEvent.RequestID
			}
		case *network.EventLoadingFinished:
			if tweetRequestId == "" {
				break
			} else {
				tweetRequestId = ""
				fc := chromedp.FromContext(ctx)
				ctx2 := cdp.WithExecutor(ctx, fc.Target)
				var searchResponse twitter.SearchResponse
				wg.Add(1)
				go func() error {
					byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
					if err != nil {
						panic(err)
					}
					err = json.Unmarshal(byts, &searchResponse)
					if err == nil {
						entries := searchResponse.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions[0].Entries
						for _, entry := range entries {
							if entry.Content.EntryType == "TimelineTimelineItem" {
								fmt.Println("Got new tweet data 😎! Saving now ....")
								result := entry.Content.ItemContent.TweetResults.Result
								tweetId := result.RestId
								userId := result.Core.Results.Result.RestId
								newTweetUrl := fmt.Sprintf("https://x.com/%s/status/%s", userId, tweetId)
								*single_tweets = append(*single_tweets, newTweetUrl)

								item := entry.Content.ItemContent
								addToTweetResult(item, tweet_results)
							}
						}
						wg.Done()
					}
					return nil
				}()
			}
		}
	})
}
