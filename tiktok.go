package main

import (
	"context"
	"encoding/json"
	"errors"
	tiktok_pkg "example/hello/pkg/tiktok"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type TweetData struct {
	auth_token  string
	tweet_url   string
	tweet_query string
}

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

var DEBUG_TIKTOK bool = true

// sesssion id
func promptTiktok() (int, error) {

	var userOption int
	if DEBUG_TIKTOK {
		userOption = singleTiktok
		tiktokData.tiktok_url = "https://www.tiktok.com/@jokilicers.id/video/7461791113637022984"
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
				firstPageUrl = preprocessURL(response.URL)
			}
		}
	})
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(tiktokData.tiktok_url),
		chromedp.WaitReady(`body #app`),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return firstPageUrl
}

var hasMore int = 1
var cursor int = 0

func handleSingleTiktok() {
	firstPageUrl := getFirstCommentPageUrl()

	ctx, acancel := createNewContext()
	defer acancel()
	tiktokListener(ctx)

	for hasMore > 0 {
		err := chromedp.Run(ctx,
			network.Enable(),
            // Todo : construct url based on has more and cursor
			chromedp.Navigate(convert(firstPageUrl, cursor)),
			chromedp.WaitVisible(`body pre`),
            chromedp.Sleep(2*time.Second),
		)
		if err != nil {
			log.Fatalln(err)
		}
	}

}

func convert(s string, newCursor int) string {
	urlResult, err := url.Parse(s)
	if err != nil {
		log.Fatalln(err)
	}
    q := urlResult.Query()
    q.Set("cursor", strconv.Itoa(newCursor))
    urlResult.RawQuery = q.Encode()
	return urlResult.String()
}

func openTiktokPage(url string) chromedp.Tasks {
	fmt.Println("Opening window...")
	tasks := chromedp.Tasks{
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitReady(`body #app`),
	}
	return tasks
}

func preprocessURL(s string) string {
	// The magic part
	res := strings.Split(s, "X-Bogus=")[0]
	res = res + ("X-Bogus")
	return res
}

func tiktokListener(ctx context.Context) {
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
					fmt.Println("Received : ")
					fmt.Println(tiktokJson)
                    hasMore = tiktokJson.HasMore
                    cursor = tiktokJson.Cursor
				}
				return nil
			}()
		}
	})
}
