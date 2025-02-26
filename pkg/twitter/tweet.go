package twitter

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/agus-wesly/GoSparrow/pkg/core"
	"github.com/agus-wesly/GoSparrow/pkg/env"
	"github.com/agus-wesly/GoSparrow/pkg/terminal"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var DEBUG bool = false

type Tweet struct {
	AuthToken    string
	Limit        int
	TweetResults map[string]TweetScrapResult

	Log *terminal.Log
}

func (t *Tweet) init() {
	t.Limit = 500
	t.TweetResults = make(map[string]TweetScrapResult)
	t.Log = &terminal.Log{}
	t.Log.NewCursor()
}

func (t *Tweet) getTokenFromUser() {
	if DEBUG {
		token, err := env.Get("TWEET_AUTH_TOKEN")
		if err != nil {
			panic(err)
		}
		t.AuthToken = token
	} else {
		inp := terminal.Input{
			Message:   "Enter your twitter auth token ",
			Validator: terminal.Required,
		}
		err := inp.Ask(&t.AuthToken)
		if err != nil {
			panic(err)
		}
	}
}

const (
	SEARCH_MODE = "Search Mode"
	SINGLE_MODE = "Single Mode"
)

func (t *Tweet) getModeFromUser() string {
	if !DEBUG {
		var res int
		s := terminal.Select{
			Opts:    []string{SEARCH_MODE, SINGLE_MODE},
			Message: "Choose Mode",
		}
		s.Ask(&res)
		return s.Opts[res]
	}
	return SINGLE_MODE
}

func (t *Tweet) Begin() {
	t.init()
	t.getTokenFromUser()
	mode := t.getModeFromUser()
	if mode == SINGLE_MODE {
		tweetSingle := TweetSingleOption{
			Tweet: t,
		}
		tweetSingle.Prompt()
		tweetSingle.BeginSingleTweet()
	} else if mode == SEARCH_MODE {
		twitterSearch := TweetSearchOption{
			Tweet:      t,
			MinReplies: 0,
			Query:      "",
			MinLikes:   0,
			Language:   "en",
		}
		twitterSearch.Prompt()
		twitterSearch.BeginSearchTweet()
	} else {
		panic("Unknown option")
	}
}

func (t *Tweet) SetupToken() {
	ctx, acancel := core.CreateNewContext()
	defer acancel()
	err := chromedp.Run(
		ctx,
		t.AttachAuthToken(),
		t.VerifyAuthToken(),
	)
	if err != nil {
		t.Log.Error(err)
		os.Exit(1)
	}
}

func (t *Tweet) AttachAuthToken() chromedp.Tasks {
	expr := cdp.TimeSinceEpoch(time.Now().Add(3 * 24 * time.Hour))
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := network.SetCookie("auth_token", t.AuthToken).
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
			return nil
		}),
	}
}

func (t *Tweet) VerifyAuthToken() chromedp.Tasks {
	var location string
	t.Log.Info("Verifying token...")
	return chromedp.Tasks{
		chromedp.Navigate("https://x.com/explore"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			act := chromedp.Location(&location)
			err := act.Do(ctx)
			if err != nil {
				return err
			}
			if strings.Contains(location, "login") {
				return errors.New("Auth token is not valid. Please provide valid auth token")
			}
			t.Log.Success("Token Verified")
			return nil
		}),
	}
}

func (t *Tweet) openTweetPage(url string) chromedp.Tasks {
	// Search for request that includes : TweetDetail
	// const url string = "https://x.com/jherr/status/1758571101964382487"
	t.Log.Info("Opening window...")
	tasks := chromedp.Tasks{
		network.Enable(),
		chromedp.Navigate(url),
		// todo : we need timeout in case this stuck
		// because the user probably not entering valid twitter url
		chromedp.WaitReady(`body [data-testid="tweetButtonInline"]`),
	}
	return tasks
}

func (t *Tweet) isReachingLimit() bool {
	currLen := len(t.TweetResults)
	return currLen >= t.Limit
}

func (t *Tweet) scroll() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight})`, nil),
		chromedp.Evaluate(`document.evaluate("//span[contains(., 'Show replies')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext()?.click()`, nil),
		chromedp.Evaluate(`document.evaluate("//button[contains(., 'Show replies')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext()?.click()`, nil),
		// todo : wait for another tweet to coming ??
		chromedp.Sleep(3 * time.Second),
	}
}

func (t *Tweet) processTweetJSON(jsonData Response) error {
	var entries []Entry
	instructions := jsonData.Data.ThreadedConversationsWithInjectionsV2.Instructions
	if len(instructions) == 0 {
		return nil
	}
	entries = instructions[0].Entries
	for i := 0; i < len(entries); i++ {
		currentEntryContent := entries[i].Content
		var item *ItemContent
		if currentEntryContent.ItemContent != nil {
			item = currentEntryContent.ItemContent
			t.addToTweetResult(item)
		}

		if currentEntryContent.Items != nil {
			items := *currentEntryContent.Items
			for j := 0; j < len(items); j++ {
				item = items[j].Item.ItemContent
				t.addToTweetResult(item)
			}
		}
	}
	return nil
}

func (t *Tweet) addToTweetResult(item *ItemContent) TweetScrapResult {
	tweet := TweetScrapResult{
		TweetId:   item.TweetResults.Result.RestId,
		Author:    item.TweetResults.Result.Core.Results.Result.Legacy.Name,
		Content:   cleanupContent(item.TweetResults.Result.Legacy.FullText),
		UserIdStr: item.TweetResults.Result.Legacy.UserIdStr,
	}
	t.TweetResults[tweet.TweetId] = tweet
	return tweet
}

func (t *Tweet) ExportToCSV() string {
	t.Log.Info("Exporting to csv...")
	res := make([][]string, len(t.TweetResults)+1)
	//TODO : Add another field to this
	res[0] = []string{"Tweet_Id", "Author", "Content"}

	var i int = 1
	for _, val := range t.TweetResults {
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
	fileName := fmt.Sprintf("res-tiktok-%d.csv", currentTime.Unix())
	os.WriteFile(fileName, buf.Bytes(), 0644)
	t.Log.Success("Successfully exported in ", fileName)
	return fileName
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

func (t *Tweet) ValidateTweetUrl(s string) bool {
	if strings.Contains(s, "x.com") || strings.Contains(s, "twitter.com") {
		if strings.Contains(s, "status") {
			return true
		}
	}
	return false
}
