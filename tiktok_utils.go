package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"example/hello/pkg/core"
	"example/hello/pkg/tiktok"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)


func getFirstCommentPageUrl() string {
	var firstPageUrl string
	ctx, acancel := core.CreateNewContext()
	defer acancel()
	// listen
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, "comment") {
				fmt.Println("Retrieved base comment, peeking all comments...")
				firstPageUrl = preprocessURL(response.URL)
			}
		}
	})
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(tiktokData.TiktokUrl),
		chromedp.WaitReady(`.css-7whb78-DivCommentListContainer`),
		// Just to make sure ...
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return firstPageUrl
}

var hasMore bool = true
var cursor int = 0

func handleSingleTiktok() {
	firstPageUrlString := getFirstCommentPageUrl()

	ctx, acancel := core.CreateNewContext()
	defer acancel()

	firstPageUrl, err := url.Parse(firstPageUrlString)
	if err != nil {
		log.Fatalln(err)
	}
	tiktokId := firstPageUrl.Query().Get("aweme_id")
	tiktokResults := make([]tiktok.TiktokScrapResult, 0)
	tiktokListener(ctx, tiktokId, &tiktokResults)

	chromedp.Run(
		ctx,
		network.Enable(),
	)

	for hasMore {
		err := chromedp.Run(ctx,
			chromedp.Navigate(updateUrl(firstPageUrlString, cursor)),
			chromedp.WaitVisible(`body pre`),
			chromedp.Sleep(2*time.Second),
		)
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Println("Starting the export process...")
	fileName := exportTiktokToCSV(tiktokResults)
	fmt.Println("Successfully exported to : ", fileName)
}

func updateUrl(s string, newCursor int) string {
	urlResult, err := url.Parse(s)
	if err != nil {
		log.Fatalln(err)
	}
	q := urlResult.Query()
	q.Set("cursor", strconv.Itoa(newCursor))
	urlResult.RawQuery = q.Encode()
	fmt.Println(urlResult.String())
	return urlResult.String()
}

func preprocessURL(s string) string {
	// The magic part
	res := strings.Split(s, "X-Bogus=")[0]
	res = res + ("X-Bogus")
	return res
}

func tiktokListener(ctx context.Context, tiktokId string, tiktokResults *[]tiktok.TiktokScrapResult) {
	// Listen phase
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventLoadingFinished:
			fc := chromedp.FromContext(ctx)
			ctx2 := cdp.WithExecutor(ctx, fc.Target)
			go func() error {
				var tiktokJson tiktok.Response
				byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
				if err != nil {
					log.Fatalln(err)
				}
				err = json.Unmarshal(byts, &tiktokJson)
				if err == nil {
					tiktokResultChunk := processTiktokJSON(tiktokJson, tiktokId)
					*tiktokResults = append(*tiktokResults, tiktokResultChunk...)
					fmt.Println("Successfully retrieved one chunk of comment...")
					hasMore = len(tiktokJson.Comments) > 0
					cursor = tiktokJson.Cursor
				}
				return nil
			}()
		}
	})
}

func processTiktokJSON(tiktokJson tiktok.Response, tiktokId string) []tiktok.TiktokScrapResult {
	var res []tiktok.TiktokScrapResult
	for i := 0; i < len(tiktokJson.Comments); i++ {
		comment := tiktokJson.Comments[i]
		res = append(res, tiktok.TiktokScrapResult{TiktokId: tiktokId, Content: comment.Text, Author: comment.User.Nickname, UserIdStr: comment.User.UniqueId})
	}
	return res
}


func exportTiktokToCSV(tiktok_results []tiktok.TiktokScrapResult) string {
	res := make([][]string, len(tiktok_results)+1)
	res[0] = []string{"Tiktok_ID", "Author", "Comment"}

	var i int = 1
	for _, val := range tiktok_results {
		res[i] = []string{val.TiktokId, val.Author, val.Content}
		i++
	}
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.WriteAll(res)
	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}

	currentTime := time.Now().Local()
	fileName := fmt.Sprintf("res-tiktok%d.csv", currentTime.Unix())
	os.WriteFile(fileName, buf.Bytes(), 0644)
	return fileName
}
