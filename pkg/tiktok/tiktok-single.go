package tiktok

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/agus-wesly/GoSparrow/pkg/core"
	"github.com/agus-wesly/GoSparrow/pkg/terminal"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type TiktokSingleOption struct {
	*Tiktok
	TiktokUrl string
	HasMore   bool
	Cursor    int

	FirstCommentUrl *url.URL
	TiktokId        string
}

var REACHING_LIMIT_ERR = errors.New("Reaching limit")

func (t *TiktokSingleOption) Prompt() {
	t.TiktokUrl = "https://www.tiktok.com/@andwey.kurt1/video/7467314067489721608"
	if !DEBUG {
		opt := terminal.Input{
			Message:   "Enter your desired tiktok url",
			Validator: terminal.Required,
		}
		err := opt.Ask(&t.TiktokUrl)
		if err != nil {
			panic(err)
		}
		opt = terminal.Input{
			Message:   "How many tiktok replies do you want to get ? [Default : 500]",
			Validator: terminal.IsNumber,
			Default:   "500",
		}
		err = opt.Ask(&t.Limit)
		if err != nil {
			panic(err)
		}
	}
}

func (t *TiktokSingleOption) BeginSingleTiktok() {
	// TODO: hide the cursor
	defer func() {
		_, err := t.exportResultToCSV()
		if err != nil {
			panic(err)
		}
	}()
	t.handleSingleTiktok()
	t.Log.Success("Finish scrapping. Total comments received : ", len(t.Results))
}

func (t *TiktokSingleOption) handleSingleTiktok() error {
	err := t.getFirstCommentUrl()
	if err != nil {
		return err
	}
	ctx, acancel := core.CreateNewContextWithTimeout(2 * time.Minute)
	defer acancel()

	t.Log.Success("starting to peak...")

	t.listenForReplies(ctx)
	err = chromedp.Run(
		ctx,
		network.Enable(),
	)
	if err != nil {
		return err
	}
	for t.HasMore {
		if t.isReachingLimit() {
			return REACHING_LIMIT_ERR
		}
		err = chromedp.Run(ctx,
			chromedp.Navigate(t.updateUrl()),
			chromedp.WaitVisible(`body pre`),
			chromedp.Sleep(1*time.Second),
		)
		if err != nil {
			t.Log.Error("Something gone wrong. ", err)
			return err
		}
	}
	return nil
}

func (t *TiktokSingleOption) listenForReplies(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventLoadingFinished:
			fc := chromedp.FromContext(ctx)
			ctx2 := cdp.WithExecutor(ctx, fc.Target)
			go func() error {
				var replies ResponseJson
				byts, _ := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
				err := json.Unmarshal(byts, &replies)
				if err == nil {
					t.processReplies(replies)
					t.Log.Success("Successfully retrieved one chunk of comment. Current comment : ", len(t.Results))
					t.HasMore = len(replies.Comments) > 0
					t.Cursor = replies.Cursor
				}
				return nil
			}()
		}
	})
}

func (t *TiktokSingleOption) getFirstCommentUrl() error {
	ctx, acancel := core.CreateNewContext()
	var err error = nil
	defer acancel()
	// listen
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, "comment") {
				t.Log.Success("Retrieved base comment, peeking all comments...")
				respUrl := t.preprocessURL(response.URL)
				firstPageUrl, errorParsed := url.Parse(respUrl)
				if errorParsed != nil {
					err = errorParsed
				} else {
					t.FirstCommentUrl = firstPageUrl
					t.TiktokId = firstPageUrl.Query().Get("aweme_id")
				}
			}
		}
	})
	t.Log.Info("Opening window...")
	err = chromedp.Run(ctx,
		network.Enable(),
		network.SetBlockedURLS([]string{"https://v*-webapp-prime.tiktok.com/video"}),
		chromedp.Navigate(t.TiktokUrl),
		chromedp.WaitReady(`.css-7whb78-DivCommentListContainer`),
		chromedp.Sleep(2*time.Second),
	)
	return err
}

func (t *TiktokSingleOption) preprocessURL(s string) string {
	// The magic part
	res := strings.Split(s, "X-Bogus=")[0]
	res = res + ("X-Bogus")
	return res
}

func (t *TiktokSingleOption) updateUrl() string {
	q := t.FirstCommentUrl.Query()
	q.Set("cursor", strconv.Itoa(t.Cursor))
	t.FirstCommentUrl.RawQuery = q.Encode()
	return t.FirstCommentUrl.String()
}

func (t *TiktokSingleOption) processReplies(tiktokJson ResponseJson) {
	var res []TiktokScrapResult
	for _, comment := range tiktokJson.Comments {
		res = append(res, TiktokScrapResult{
			TiktokId:  t.TiktokId,
			Content:   comment.Text,
			Author:    comment.User.Nickname,
			UserIdStr: comment.User.UniqueId,
		})
	}
	t.Tiktok.Results = append(t.Tiktok.Results, res...)
}
