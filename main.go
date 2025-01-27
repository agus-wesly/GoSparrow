package main

import (
	"context"
	"encoding/json"
	"fmt"
	// "log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Response struct {
	Data Data `json:"data"`
}

func main() {
	data, err := os.ReadFile("tweet-response.json")
	if err != nil {
		panic(err)
	}

	var response Response
	json.Unmarshal(data, &response)
	processTweetJSON(response)

	return
	// create instance

	// ctx, cancel := chromedp.NewContext(context.Background())
	// defer cancel()

	// ctx, cancel = context.WithTimeout(ctx, 100*time.Second)
	// defer cancel()

	// // navigate to a page
	// err := chromedp.Run(ctx,
	// 	authenticate("auth_token", "c9bca772a8e05e076c17da20f126d22e042dae6b"),
	// 	verifyLogin(),
	// 	visitPageAndDownload(ctx),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func visitPageAndDownload(ctx context.Context) chromedp.Tasks {
	var tweetRequestId network.RequestID = ""
	// Search for request that includes : TweetDetail
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
				fmt.Println(responseReceivedEvent.RequestID)
				tweetRequestId = ""
				fc := chromedp.FromContext(ctx)
				ctx2 := cdp.WithExecutor(ctx, fc.Target)
				var tweetJson interface{}
				go func() {
					byts, err := network.GetResponseBody(responseReceivedEvent.RequestID).Do(ctx2)
					if err != nil {
						panic(err)
					}
					json.Unmarshal(byts, &tweetJson)
					// saveToJsonFile(byts)
					// processTweetJSON(&tweetJson)
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

func saveToJsonFile(data []byte) {
	err := os.WriteFile("tweet-response.json", data, 0644)
	if err != nil {
		panic(err)
	}
}

type Data struct {
	ThreadedConversationsWithInjectionsV2 ThreadedConversation `json:"threaded_conversation_with_injections_v2"`
}

type ThreadedConversation struct {
	Instructions []Instruction `json:"instructions"`
}

type Instruction struct {
	Type    string  `json:"type"`
	Entries []Entry `json:"entries"`
}

type Entry struct {
	EntryId   string  `json:"entryId"`
	SortIndex string  `json:"sortIndex"`
	Content   Content `json:"content"`
}

type Content struct {
	EntryType   string       `json:"entryType"`
	ItemContent *ItemContent `json:"itemContent"`
	Items       *[]Items     `json:"items"`
}

type Items struct {
	EntryId string `json:"entryId"`
	Item    Item   `json:"item"`
}

type Item struct {
	ItemType     string       `json:"itemType"`
	TweetResults TweetResults `json:"tweet_results"`
	ItemContent  *ItemContent `json:"itemContent"`
}

type ItemContent struct {
	ItemType     string       `json:"itemType"`
	TweetResults TweetResults `json:"tweet_results"`
}

type TweetResults struct {
	Result Result `json:"result"`
}

type Result struct {
	RestId string `json:"rest_id"`
	Legacy Legacy `json:"legacy"`
}

type Legacy struct {
	FullText string `json:"full_text"`
}

func processTweetJSON(jsonData Response) {
	var entries []Entry
	// TODO : Store the results inside this array
	// var results []string
	entries = jsonData.Data.ThreadedConversationsWithInjectionsV2.Instructions[0].Entries
	for i := 0; i < len(entries); i++ {
		currentEntryContent := entries[i].Content
		var item *ItemContent
		if currentEntryContent.ItemContent != nil {
			item = currentEntryContent.ItemContent
		}

		if currentEntryContent.Items != nil {
			items := *currentEntryContent.Items
			for j := 0; j < len(items); j++ {
				item = items[j].Item.ItemContent
			}
		}
		fmt.Println(item.TweetResults.Result.Legacy.FullText)
	}
}
