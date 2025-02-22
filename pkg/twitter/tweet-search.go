package twitter

import (
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
	*Tweet
	Query      string
	MinReplies int
	MinLikes   int
	Language   string
}

func (t *TweetSearchOption) Prompt() {
	if !DEBUG {
		inp := terminal.Input{Message: "Enter your tweet search keyword", Validator: terminal.Required}
		if err := inp.Ask(&t.Query); err != nil {
			panic(err)
		}
		inp = terminal.Input{Message: "Minimum tweet replies (Default=0)", Default: "0", Validator: terminal.IsNumber}
		if err := inp.Ask(&t.MinReplies); err != nil {
			panic(err)
		}
		inp = terminal.Input{Message: "Minimum tweet likes (Default=0)", Default: "0", Validator: terminal.IsNumber}
		if err := inp.Ask(&t.MinLikes); err != nil {
			panic(err)
		}
		inp = terminal.Input{Message: "Language [en/id] (Default=en)", Default: "en"}
		if err := inp.Ask(&t.Language); err != nil {
			panic(err)
		}
		inp = terminal.Input{
			Message:   "How many tweets do you want to retrieve ? [Default : 500]",
			Default:   "500",
			Validator: terminal.IsNumber,
		}
		if err := inp.Ask(&t.Limit); err != nil {
			panic(err)
		}
	} else {
		t.Query = "var indonesia"
		t.MinReplies = 10
	}
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
	t.SetupToken()

	ctx, cancel := core.CreateNewContext()
	defer cancel()

	err := chromedp.Run(ctx, t.AttachAuthToken())
	if err != nil {
		panic(err)
	}

	searchUrl := t.constructSearchUrl()
	fmt.Println(searchUrl)
	var entry_list = make([]Entry, 0)
	defer func() {
		fmt.Println("Exporting to csv...")
		fileName := t.ExportToCSV()
		fmt.Println("Successfully exported at : ", fileName)
	}()

	// listening in search phase
	var wg sync.WaitGroup
	core.ListenEvent(ctx, "SearchTimeline", func(byts []byte) {
		go func() error {
			var searchResponse SearchResponse
			err := json.Unmarshal(byts, &searchResponse)
			if err == nil {
				entries := searchResponse.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions[0].Entries
				for _, entry := range entries {
					if entry.Content.EntryType == "TimelineTimelineItem" {
						entry_list = append(entry_list, entry)
					}
				}
				fmt.Printf("Done searching, found %d data \n", len(entries))
				wg.Done()
				return nil
			}
			return nil
		}()
	}, &wg)
	fmt.Printf("Searching for tweets...\n")
	err = chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(searchUrl),
		chromedp.WaitReady(`body [aria-label='Timeline: Search timeline']`),
	)
	if err != nil {
		log.Fatalln(err)
	}

	wg.Wait()
	// Listen for incoming tweet
	core.ListenEvent(ctx, "TweetDetail", func(byts []byte) {
		var tweetJson Response
		err = json.Unmarshal(byts, &tweetJson)
		if err == nil {
			fmt.Println("Got new tweet data 😎! Saving now ....")
			err = t.processTweetJSON(tweetJson)
		}
	}, nil)
	for _, entry := range entry_list {
		item := entry.Content.ItemContent
		t.addToTweetResult(item)

		result := item.TweetResults.Result
		tweetId := result.RestId
		userId := result.Core.Results.Result.RestId
		newTweet := TweetSingleOption{
			Tweet:    t.Tweet,
			TweetUrl: fmt.Sprintf("https://x.com/%s/status/%s", userId, tweetId),
			Context:  ctx,
		}
		newTweet.handleSingleTweet()
	}
	// Todo : After this, figure out how to search again if there is still search limit
}
