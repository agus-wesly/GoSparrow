package twitter

import (
	"context"
	"encoding/json"
	"example/hello/pkg/core"
	"example/hello/pkg/terminal"
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type TweetSearchOption struct {
	Tweet
	Query      string
	MinReplies int
	MinLikes   int
	Language   string
}

func (t *TweetSearchOption) Prompt() {

	if !DEBUG {
		inp := terminal.Input{Message: "Enter your tweet search keyword : ", Validator: terminal.Required}
		err := inp.Ask(&t.Query)
		if err != nil {
			panic(err)
		}
		inp = terminal.Input{Message: "Minimum tweet replies (Default=0) : ", Default: "0", Validator: terminal.IsNumber}
		inp.Ask(&t.MinReplies)
		inp = terminal.Input{Message: "Minimum tweet likes (Default=0) : ", Default: "0", Validator: terminal.IsNumber}
		inp.Ask(&t.MinLikes)
		inp = terminal.Input{Message: "Language [en/id] (Default=en) : ", Default: "en"}
		inp.Ask(&t.Language)
		inp = terminal.Input{Message: "How many tweets do you want to retrieve ? [Default : 500] : ", Default: "500", Validator: terminal.IsNumber}
		inp.Ask(&t.Limit)
	} else {
		t.Query = "var indonesia"
		t.MinReplies = 10
	}
	fmt.Println(t)
}

func (t *TweetSearchOption) constructSearchUrl() string {
	parsed, err := url.Parse("https://x.com/search")
	if err != nil {
		log.Fatalln(err)
	}
	query := parsed.Query()
	res := fmt.Sprintf("%s min_replies:%d min_faves:%d lang:%s", t.Query, t.MinReplies, t.MinLikes, t.Language)
	query.Set("q", res)
	query.Set("src", "typed_query")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func (t *TweetSearchOption) BeginSearchTweet() {
	ctx, cancel := core.CreateNewContext()
	// ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	searchUrl := t.constructSearchUrl()
	fmt.Println(searchUrl)
	// var tweetData.TweetResults = make(map[string]twitter.TweetScrapResult)
	var entry_list = make([]Entry, 0)
	defer func() {
		fmt.Println("Exporting to csv...")
		fileName := t.ExportToCSV()
		fmt.Println("Successfully exported at : ", fileName)
	}()

	// listening in search phase
	var wg sync.WaitGroup
	t.Listener(ctx, "SearchTimeline", func(requestId network.RequestID, ctx2 context.Context) {
		wg.Add(1)
		go func() error {
			byts, err := network.GetResponseBody(requestId).Do(ctx2)
			if err != nil {
				return err
			}
			var searchResponse SearchResponse
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
		t.AttachAuthToken(),
		network.Enable(),
		chromedp.Navigate(searchUrl),
		chromedp.WaitReady(`body [aria-label='Timeline: Search timeline']`),
	)
	// Todo
	// scrollUntilBottom(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	wg.Wait()
	// Listen for incoming tweet
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
				if err != nil {
					return nil
				}
			}
			return nil
		}()
	})
	for _, entry := range entry_list {
		item := entry.Content.ItemContent
		t.addToTweetResult(item)

		result := item.TweetResults.Result
		tweetId := result.RestId
		userId := result.Core.Results.Result.RestId
		newTweet := TweetSingleOption{
			TweetUrl: fmt.Sprintf("https://x.com/%s/status/%s", userId, tweetId),
			Context:  ctx,
		}
		newTweet.handleSingleTweet()
	}
	// Todo : After this, figure out how to search again if there is still search limit
}
