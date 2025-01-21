package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	// create instance
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	// navigate to a page
	err := chromedp.Run(ctx,
		authenticate("auth_token", "c9bca772a8e05e076c17da20f126d22e042dae6b"),
		verifyLogin(),
		visitPageAndDownload(ctx),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func visitPageAndDownload(ctx context.Context) chromedp.Tasks {
	var evens []network.RequestID
	// Search for request that includes : TweetDetail
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if strings.Contains(response.URL, "TweetDetail") {
				evens = append(evens, responseReceivedEvent.RequestID)
			}
		case *network.EventLoadingFinished:
			i := slices.Index(evens, responseReceivedEvent.RequestID)
			if i < 0 {
				break
			} else {
				fmt.Println(evens)
				fmt.Println(responseReceivedEvent.RequestID)
				evens = slices.Delete(evens, i, i+1)
				fc := chromedp.FromContext(ctx)
				ctx2 := cdp.WithExecutor(ctx, fc.Target)
				go func() {
					byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
					if err != nil {
						panic(err)
					}
                    // Todo : change to json
					fmt.Println(byts)
				}()
				fmt.Println("OK")
			}
		}
	})
	const url string = "https://x.com/Indostransfer/status/1769976105468104944"
	return chromedp.Tasks{
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitReady(`body [data-testid="tweetButtonInline"]`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("called")
			return nil
		}),
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
			fmt.Println(location)
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
