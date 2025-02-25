package twitter

import (
	"encoding/json"
	"errors"
	"github.com/agus-wesly/GoSparrow/pkg/core"
	"github.com/agus-wesly/GoSparrow/pkg/terminal"
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

	entryList []Entry
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
	t.entryList = make([]Entry, 0)
	t.SetupToken()
    err := t.startSearching()
    if err != nil {
        t.Log.Error(err.Error())
        return
    }
	t.startScraping()
}

func (t *TweetSearchOption) startScraping() {
	defer t.ExportToCSV()
	ctx, cancel := core.CreateNewContextWithTimeout(5 * time.Minute)
	defer cancel()
	chromedp.Run(ctx, t.AttachAuthToken())
	// Listen for incoming tweet
	core.ListenEvent(ctx, "TweetDetail", func(byts []byte) {
		var tweetJson Response
		err := json.Unmarshal(byts, &tweetJson)
		if err == nil {
			t.Log.Success("Got new replies 😎! Current Replies : ", len(t.TweetResults))
			err = t.processTweetJSON(tweetJson)
		}
	}, nil)
	for _, entry := range t.entryList {
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
			if err == REACHING_LIMIT_ERR {
				break
			} else {
				t.Log.Error(err)
				continue
			}
		}
	}
	t.Log.Success("Finish scrapping. Total tweet received : ", len(t.TweetResults))
	// Todo : After this, figure out how to search again if there is still search limit
}

func (t *TweetSearchOption) startSearching() error {
	ctx, cancel := core.CreateNewContext()
	err := chromedp.Run(ctx, t.AttachAuthToken())
	if err != nil {
		panic(err)
	}
	searchUrl := t.constructSearchUrl()
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
						t.entryList = append(t.entryList, entry)
					}
				}
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
		chromedp.WaitVisible(`body [aria-label='Timeline: Search timeline'],body [data-testid='empty_state_header_text']`),
	)
	if err != nil {
		panic(err)
	}
	cancel()
	wg.Wait()
	if len(t.entryList) == 0 {
		return errors.New("No related tweet found. You may want to change your search query")
	} else {
		t.Log.Success("Done searching, found ", len(t.entryList), " related tweet")
	}
    return nil
}
