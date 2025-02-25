package tiktok

import (
	"context"
	"encoding/json"
	"errors"
	"example/hello/pkg/core"
	"example/hello/pkg/terminal"
	"fmt"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type TiktokSearchOption struct {
	*Tiktok
	Query             string
	RelatedVideos     []string
	videoListResponse VideoListResponse
}

func (t *TiktokSearchOption) init() {
	t.RelatedVideos = make([]string, 0)
}

// api/search/general/full
// tiktok.com/@${item.author.uniqueId}/video/${item.video.id}
func (t *TiktokSearchOption) Prompt() {
	t.init()
	t.Query = "penerapan var di indonesia"
	if !DEBUG {
		inp := terminal.Input{
			Message:   "Enter your tiktok search keyword",
			Validator: terminal.Required,
		}
		err := inp.Ask(&t.Query)
		if err != nil {
			panic(err)
		}
		inp = terminal.Input{
			Message:   "How many tiktok replies do you want to get ? [Default : 500]",
			Validator: terminal.IsNumber,
			Default:   "500",
		}
		err = inp.Ask(&t.Limit)
		if err != nil {
			panic(err)
		}
	}
}
func (t *TiktokSearchOption) BeginSearchTiktok() {
	defer t.Tiktok.exportResultToCSV()
	err := t.searchRelevantVideo()
	if err != nil {
		t.Log.Error(err)
	}
	t.processEachVideo()
}

func (t *TiktokSearchOption) searchRelevantVideo() error {
	// https://www.tiktok.com/search?q=penerapan%20var%20di%20indonesia
	t.Log.Info("Searching relevant video...")
	ctx, cancel := core.CreateNewContext()
	defer cancel()
	searchUrl := t.constructUrl()

	core.ListenEvent(ctx, "api/search/general/full", func(byts []byte) {
		err := json.Unmarshal(byts, &t.videoListResponse)
		if err == nil {
			t.processVideoList()
		}
	}, nil)

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(searchUrl),
		chromedp.WaitVisible(`body .css-1soki6-DivItemContainerForSearch`),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		return err
	}
	if len(t.RelatedVideos) == 0 {
		return errors.New("Cannot get any related video. Please change your search query")
	}

	t.Log.Success("Successfully get ", len(t.RelatedVideos), " of related videos")
	return nil
}

func (t *TiktokSearchOption) processEachVideo() {
	for _, videoUrl := range t.RelatedVideos {
		tiktokVideo := TiktokSingleOption{
			Tiktok:    t.Tiktok,
			TiktokUrl: videoUrl,
			HasMore:   true,
		}
		err := tiktokVideo.handleSingleTiktok()
		if err != nil {
			if err == REACHING_LIMIT_ERR {
				break
			} else {
				continue
			}
		}
	}
	t.Log.Success("Finish scrapping. Total comments received : ", len(t.Results))
}

func (t *TiktokSearchOption) constructUrl() string {
	baseUrl := "https://www.tiktok.com/search"
	parsed, err := url.Parse(baseUrl)
	if err != nil {
		panic(err)
	}
	q := parsed.Query()
	q.Add("q", t.Query)
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

type ExecFn func(requestId network.RequestID, ctx2 context.Context)

func (t *TiktokSearchOption) processVideoList() {
	// tiktok.com/@${item.author.uniqueId}/video/${item.video.id}
	if t.videoListResponse.Data != nil {
		for _, response := range t.videoListResponse.Data {
			authorId := response.Item.Author.UniqueId
			videoId := response.Item.Video.Id
			if authorId == "" || videoId == "" {
				continue
			}
			newRelatedVideo := fmt.Sprintf("https://tiktok.com/@%s/video/%s", authorId, videoId)
			t.RelatedVideos = append(t.RelatedVideos, newRelatedVideo)
		}
	}
}
