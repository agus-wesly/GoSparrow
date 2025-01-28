package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	// "log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var tweet_results = make(map[string]TweetScrapResult)

func main() {
	// Disable the headless mode to see what happen.
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()

	// create instance
	ctx, cancel := chromedp.NewContext(actx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	// Listen phase
	var tweetRequestId network.RequestID = ""
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, "TweetDetail") {
				tweetRequestId = responseReceivedEvent.RequestID
			}
		case *network.EventLoadingFinished:
			if tweetRequestId == "" {
				break
			} else {
				tweetRequestId = ""
				fc := chromedp.FromContext(ctx)
				ctx2 := cdp.WithExecutor(ctx, fc.Target)
				var tweetJson Response
				go func() {
					byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
					if err != nil {
						panic(err)
					}
					json.Unmarshal(byts, &tweetJson)
					fmt.Println("Got bytes !")
					// saveToJsonFile(byts)
					// processTweetJSON(tweetJson)
				}()
			}
		}
	})

	err := chromedp.Run(ctx,
		authenticate("auth_token", "c9bca772a8e05e076c17da20f126d22e042dae6b"),
		verifyLogin(),
		openPage(),
	)

	if err != nil {
		log.Fatalln(err)
	}

    // TODO : stop scroll when reach bottom of the page
	for i := 0; i < 3; i++ {
		chromedp.Run(ctx, scrollDown())
	}
}

func openPage() chromedp.Tasks {
	// Search for request that includes : TweetDetail
	const url string = "https://x.com/Indostransfer/status/1769976105468104944"
	tasks := chromedp.Tasks{
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitReady(`body [data-testid="tweetButtonInline"]`),
	}
	return tasks
}

func scrollDown() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight})`, nil),
		chromedp.Evaluate(`document.querySelectorAll("a div[data-testid='tweetPhoto']").forEach((el) => el.remove())`, nil),
		chromedp.Sleep(5 * time.Second),
	}
}

func verifyLogin() chromedp.Tasks {
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
				panic("Invalid auth token")
			}
			return nil
		}),
	}

}

func authenticate(cookies ...string) chromedp.Tasks {
	if len(cookies)%2 != 0 {
		panic("Length must be divisible by 2")
	}
	expr := cdp.TimeSinceEpoch(time.Now().Add(3 * 24 * time.Hour))
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < len(cookies)-1; i++ {
				err := network.SetCookie(cookies[i], cookies[i+1]).
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
			}
			return nil
		}),
	}

}
