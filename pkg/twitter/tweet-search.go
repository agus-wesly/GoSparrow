package twitter

import (
	"encoding/json"
	"example/hello/pkg/core"
	"example/hello/pkg/terminal"
	"fmt"
	"net/url"
	"sync"
	"time"

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
		panic(err)
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

	err := chromedp.Run(ctx, t.AttachAuthToken())
	if err != nil {
		panic(err)
	}

	searchUrl := t.constructSearchUrl()
	var entry_list = make([]Entry, 0)
	defer func() {
		t.ExportToCSV()
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
				t.Log.Success("Done searching, found ", len(entries), " related tweet")
				wg.Done()
				return nil
			}
			return nil
		}()
	}, &wg)
	t.Log.Info("Searching for tweets...")
	err = chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(searchUrl),
		chromedp.WaitReady(`body [aria-label='Timeline: Search timeline']`),
	)
	if err != nil {
		panic(err)
	}
	cancel()

	wg.Wait()
	// create a new context ?
	ctx, cancel = core.CreateNewContextWithTimeout(2 * time.Minute)
	defer cancel()
	chromedp.Run(ctx, t.AttachAuthToken())
	// Listen for incoming tweet
	core.ListenEvent(ctx, "TweetDetail", func(byts []byte) {
		var tweetJson Response
		err = json.Unmarshal(byts, &tweetJson)
		if err == nil {
			t.Log.Success("Got new replies 😎! Current Replies : ", len(t.TweetResults))
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
			Context:  &ctx,
		}
		err := newTweet.handleSingleTweet()
		if err != nil {
			t.Log.Error(err)
			continue
		}
	}
	// Todo : After this, figure out how to search again if there is still search limit
}
