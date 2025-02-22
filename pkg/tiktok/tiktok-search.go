package tiktok

import (
	"context"
	"encoding/json"
	"example/hello/pkg/core"
	"example/hello/pkg/terminal"
	"fmt"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type TiktokSearchOption struct {
	Tiktok            *Tiktok
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
	}
}
func (t *TiktokSearchOption) BeginSearchTiktok() {
	defer t.Tiktok.exportResultToCSV()
	t.searchRelevantVideo()
	t.processEachVideo()
}

func (t *TiktokSearchOption) searchRelevantVideo() {
	// https://www.tiktok.com/search?q=penerapan%20var%20di%20indonesia
	ctx, cancel := core.CreateNewContext()
	defer cancel()
	searchUrl := t.constructUrl()

	core.ListenEvent(ctx, "api/search/general/full", func(byts []byte) {
        err := json.Unmarshal(byts, &t.videoListResponse)
		if err == nil {
			fmt.Println("Got tiktok video 😎! Saving now ....")
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
		panic(err)
	}
	fmt.Println(t.RelatedVideos)
}

func (t *TiktokSearchOption) processEachVideo() {
	for _, videoUrl := range t.RelatedVideos[0:3] {
		tiktokVideo := TiktokSingleOption{
			Tiktok:    t.Tiktok,
			TiktokUrl: videoUrl,
			HasMore:   true,
		}
		tiktokVideo.handleSingleTiktok()
	}
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
