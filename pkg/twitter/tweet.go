package twitter

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"example/hello/pkg/core"
	"example/hello/pkg/terminal"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var DEBUG bool = false

type Tweet struct {
	AuthToken    string
	Limit        int
	TweetResults map[string]TweetScrapResult
}

func (t *Tweet) init() {
	t.Limit = 500
	t.TweetResults = make(map[string]TweetScrapResult)
}

func (t *Tweet) getTokenFromUser() {
	if DEBUG {
		t.AuthToken = "c9bca772a8e05e076c17da20f126d22e042dae6b"
	} else {
		fmt.Print("Enter your twitter auth token : ")
		fmt.Scanln(&t.AuthToken)
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
			Tweet: *t,
		}
		tweetSingle.Prompt()
	} else if mode == SEARCH_MODE {
		twitterSearch := TweetSearchOption{
			Tweet:      *t,
			MinReplies: 0,
			Query:      "",
			MinLikes:   0,
			Language:   "en",
		}
		twitterSearch.Prompt()
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
		fmt.Println("Auth token is invalid", err)
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

func openTweetPage(url string) chromedp.Tasks {
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

func (t *Tweet) isReachingLimit() bool {
	currLen := len(t.TweetResults)
	fmt.Println("Current length : ", currLen)
	return currLen >= t.Limit
}

func (t *Tweet) scroll() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight})`, nil),
		chromedp.Evaluate(`document.evaluate("//span[contains(., 'Show replies')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext()?.click()`, nil),
		chromedp.Evaluate(`document.evaluate("//button[contains(., 'Show replies')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext()?.click()`, nil),
		chromedp.Sleep(2 * time.Second),
	}
}

func (t *Tweet) scrollUntilBottom(ctx context.Context) {
	var isAlreadyOnTheBottom bool = false
	for !isAlreadyOnTheBottom && !t.isReachingLimit() {
		err := chromedp.Run(ctx, t.scroll())
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

type ExecFn func(requestId network.RequestID, ctx2 context.Context)

func (t *Tweet) Listener(ctx context.Context, responseKey string, exec ExecFn) error {
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
	res := make([][]string, len(t.TweetResults)+1)
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
	fileName := fmt.Sprintf("res-%d.csv", currentTime.Unix())
	os.WriteFile(fileName, buf.Bytes(), 0644)
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
