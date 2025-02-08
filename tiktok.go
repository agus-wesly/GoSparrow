package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	tiktok_pkg "example/hello/pkg/tiktok"
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

const (
	searchTiktok = iota + 1
	singleTiktok
)

var global_url_to_visit = make([]string, 0)

type TiktokData struct {
	tiktok_url   string
	tiktok_query string
}

var tiktokData TiktokData

var DEBUG_TIKTOK bool = false

// sesssion id
func promptTiktok() (int, error) {

	var userOption int
	if DEBUG_TIKTOK {
		userOption = singleTiktok
		tiktokData.tiktok_url = "https://www.tiktok.com/@andwey.kurt1/video/7467314067489721608"
	}
	if !DEBUG_TIKTOK {
		fmt.Println("=====CHOOSE MODE=====")
		fmt.Println("1. Search Mode")
		fmt.Println("2. Single tiktok Mode")
		fmt.Print("Your Option : ")
		fmt.Scan(&userOption)
	}

	if userOption == searchTiktok {
		return 0, errors.New("Search mode is not yet implemented")
	} else if userOption == singleTiktok {
		if !DEBUG_TIKTOK {
			fmt.Print("Enter your desired tiktok url : ")
			fmt.Scanln(&tiktokData.tiktok_url)
		}
		return singleTiktok, nil
	} else {
		return 0, errors.New("Option provided is unknown")
	}

}

func getFirstCommentPageUrl() string {
	var firstPageUrl string
	ctx, acancel := createNewContext()
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
		chromedp.Navigate(tiktokData.tiktok_url),
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

	ctx, acancel := createNewContext()
	defer acancel()

	firstPageUrl, err := url.Parse(firstPageUrlString)
	if err != nil {
		log.Fatalln(err)
	}
	tiktokId := firstPageUrl.Query().Get("aweme_id")
	tiktokResults := make([]TiktokScrapResult, 0)
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

func tiktokListener(ctx context.Context, tiktokId string, tiktokResults *[]TiktokScrapResult) {
	// Listen phase
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventLoadingFinished:
			fc := chromedp.FromContext(ctx)
			ctx2 := cdp.WithExecutor(ctx, fc.Target)
			go func() error {
				var tiktokJson tiktok_pkg.Response
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

func processTiktokJSON(tiktokJson tiktok_pkg.Response, tiktokId string) []TiktokScrapResult {
	var res []TiktokScrapResult
	for i := 0; i < len(tiktokJson.Comments); i++ {
		comment := tiktokJson.Comments[i]
		res = append(res, TiktokScrapResult{TiktokId: tiktokId, Content: comment.Text, Author: comment.User.Nickname, UserIdStr: comment.User.UniqueId})
	}
	return res
}

type TiktokScrapResult struct {
	TiktokId  string `json:"tiktok_id"`
	Author    string `json:"username"`
	Content   string `json:"content"`
	UserIdStr string `json:"user_id_str"`
}

func exportTiktokToCSV(tiktok_results []TiktokScrapResult) string {
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
