package twitter

import (
	"bufio"
	"context"
	"encoding/json"
	"example/hello/pkg/core"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
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
		fmt.Print("Enter your tweet search keyword : ")
		in := bufio.NewReader(os.Stdin)
		inputQuery, err := in.ReadString('\n')
		if err != nil {
			log.Fatalln(err)
		}
		// Todo : handle default value and retry mechanism
		inputQuery = strings.TrimSpace(inputQuery)
		t.Query = inputQuery
		fmt.Print("Minimum tweet replies (Default=0) : ")
		fmt.Scan(&t.MinReplies)
		fmt.Print("Minimum tweet likes (Default=0) : ")
		fmt.Scan(&t.MinLikes)
		fmt.Print("Language [en/id] (Default=en) : ")
		fmt.Scan(&t.Language)
		fmt.Print("How many tweets do you want to retrieve ? [Default : 500] : ")
		fmt.Scanln(&t.Limit)
	} else {
		t.Query = "var indonesia"
		t.MinReplies = 10
	}
    t.beginSearchTweet()
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

func (t *TweetSearchOption) beginSearchTweet() {
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
	wg.Wait()
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
		newTweet.BeginSingleTweet()
	}
	// Todo : After this, figure out how to search again if there is still search limit
}
