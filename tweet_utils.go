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

type ExecFn func(requestId network.RequestID, ctx2 context.Context)

func tweetListener(ctx context.Context, responseKey string, exec ExecFn) error {
	tweetRequestIdList := make([]network.RequestID, 0)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, responseKey) {
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
				exec(responseReceivedEvent.RequestID, ctx2)
			}
		}
	})
	return nil
}

func exec(requestId network.RequestID, ctx2 context.Context) error {
	byts, err := network.GetResponseBody(requestId).Do(ctx2)
	if err != nil {
		fmt.Println("No resource error")
		return err
	}
	var tweetJson twitter.Response
	err = json.Unmarshal(byts, &tweetJson)
	if err == nil {
		fmt.Println("Got new tweet data 😎! Saving now ....")
		processTweetJSON(tweetJson)
	}
	return nil
}

func _tweetListener(ctx context.Context) {
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
						return err
					}
					err = json.Unmarshal(byts, &tweetJson)
					if err == nil {
						fmt.Println("Got new tweet data 😎! Saving now ....")
						processTweetJSON(tweetJson)
					}
					return nil
				}()
			}
		}
	})
}

func isReachingLimit() bool {
	currLen := len(tweetData.TweetResults)
	fmt.Println("Current length : ", currLen)
	return currLen >= tweetData.Limit
}

func scrollUntilBottom(ctx context.Context) {
	var isAlreadyOnTheBottom bool = false
	for !isAlreadyOnTheBottom && !isReachingLimit() {
		err := chromedp.Run(ctx, scrollTweetDown())
		if err != nil {
			break
		}
		chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("Scrolling down...")
			return nil
		}))
		chromedp.Run(ctx, chromedp.Evaluate(`Math.round(window.scrollY) + window.innerHeight >= document.body.scrollHeight`, &isAlreadyOnTheBottom))
	}
}

func tweetActions(ctx context.Context, url string) error {
	err := chromedp.Run(ctx,
		openPage(url),
	)
	if err != nil {
		return err
	}
	scrollUntilBottom(ctx)
	return nil
}

func handleSingleTweet(ctx context.Context, tweetUrl string) {
	err := tweetActions(ctx, tweetUrl)
	if err != nil {
		fmt.Println("Error : ", err)
	}
}

func beginSingleTweet(tweetUrl string) {

	defer func() {
		fmt.Println("Exporting to csv...")
		fileName := exportTweetToCSV()
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
	tweetListener(ctx, "TweetDetail", func(requestId network.RequestID, ctx2 context.Context) {
		go func() error {
			byts, err := network.GetResponseBody(requestId).Do(ctx2)
			if err != nil {
				fmt.Println("No resource error")
				return err
			}
			var tweetJson twitter.Response
			err = json.Unmarshal(byts, &tweetJson)
			if err == nil {
				fmt.Println("Got new tweet data 😎! Saving now ....")
				err = processTweetJSON(tweetJson)
			}
			return nil
		}()
	},
	)
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

func processTweetJSON(jsonData twitter.Response) error {
	var entries []twitter.Entry
	instructions := jsonData.Data.ThreadedConversationsWithInjectionsV2.Instructions
	if len(instructions) == 0 {
		return nil
	}
	entries = instructions[0].Entries
	for i := 0; i < len(entries); i++ {
		currentEntryContent := entries[i].Content
		var item *twitter.ItemContent
		if currentEntryContent.ItemContent != nil {
			item = currentEntryContent.ItemContent
			addToTweetResult(item)
		}

		if currentEntryContent.Items != nil {
			items := *currentEntryContent.Items
			for j := 0; j < len(items); j++ {
				item = items[j].Item.ItemContent
				addToTweetResult(item)
			}
		}
	}
	return nil
}

func cleanupContent(content string) string {
	if content == "" || content[0] != '@' {
		return content
	}
	idx := 0
	for idx < len(content) && content[idx] != ' ' {
		idx++
	}
	res := strings.Trim(content[idx:], " ")
	return cleanupContent(res)
}

func addToTweetResult(item *twitter.ItemContent) twitter.TweetScrapResult {
	tweet := twitter.TweetScrapResult{
		TweetId:   item.TweetResults.Result.RestId,
		Author:    item.TweetResults.Result.Core.Results.Result.Legacy.Name,
		Content:   cleanupContent(item.TweetResults.Result.Legacy.FullText),
		UserIdStr: item.TweetResults.Result.Legacy.UserIdStr,
	}
	tweetData.TweetResults[tweet.TweetId] = tweet
	return tweet
}

func exportTweetToCSV() string {
	res := make([][]string, len(tweetData.TweetResults)+1)
	res[0] = []string{"Tweet_Id", "Author", "Content"}

	var i int = 1
	for _, val := range tweetData.TweetResults {
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

func scrollTweetDown() chromedp.Tasks {
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

	searchUrl := constructSearchUrl(opt)
	fmt.Println(searchUrl)
	// var tweetData.TweetResults = make(map[string]twitter.TweetScrapResult)
	var entry_list = make([]twitter.Entry, 0)
	defer func() {
		fmt.Println("Exporting to csv...")
		fileName := exportTweetToCSV()
		fmt.Println("Successfully exported at : ", fileName)
	}()

	// listening in search phase
	var wg sync.WaitGroup
	tweetListener(ctx, "SearchTimeline", func(requestId network.RequestID, ctx2 context.Context) {
		wg.Add(1)
		go func() error {
			byts, err := network.GetResponseBody(requestId).Do(ctx2)
			if err != nil {
				return err
			}
			var searchResponse twitter.SearchResponse
			err = json.Unmarshal(byts, &searchResponse)
			if err == nil {
				entries := searchResponse.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions[0].Entries
				for _, entry := range entries {
					if entry.Content.EntryType == "TimelineTimelineItem" {
						fmt.Println("Got new tweet data 😎! Saving now ....")
						entry_list = append(entry_list, entry)
					}
				}
				wg.Done()
				return nil
			}
			return nil
		}()
	})
	err := chromedp.Run(ctx,
		authenticateTwitter("auth_token", tweetData.AuthToken),
		network.Enable(),
		chromedp.Navigate(searchUrl),
		chromedp.WaitReady(`body [aria-label='Timeline: Search timeline']`),
	)
	// Todo
	// scrollUntilBottom(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	// Listen for incoming tweet
	tweetListener(ctx, "TweetDetail", func(requestId network.RequestID, ctx2 context.Context) {
		go func() error {
			byts, err := network.GetResponseBody(requestId).Do(ctx2)
			if err != nil {
				fmt.Println("No resource error")
				return err
			}

			var tweetJson twitter.Response
			err = json.Unmarshal(byts, &tweetJson)
			if err == nil {
				fmt.Println("Got new tweet data 😎! Saving now ....")
				err = processTweetJSON(tweetJson)
				if err != nil {
					return nil
				}
			}
			return nil
		}()
	})
	wg.Wait()
	for _, entry := range entry_list {
		result := entry.Content.ItemContent.TweetResults.Result
		tweetId := result.RestId
		userId := result.Core.Results.Result.RestId
		item := entry.Content.ItemContent
		addToTweetResult(item)
		newTweetUrl := fmt.Sprintf("https://x.com/%s/status/%s", userId, tweetId)
		handleSingleTweet(ctx, newTweetUrl)
	}
	// Todo : After this, figure out how to search again if there is still search limit
}
