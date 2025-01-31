package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	searchTweet = iota + 1
	singleTweet
)

type TweetData struct {
	auth_token  string
	tweet_url   string
	tweet_query string
}

var tweetData TweetData

var DEBUG bool = true

func main() {
	opt, err := promptUser()
	if err != nil {
		log.Fatalln(err)
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()
	ctx, cancel := chromedp.NewContext(actx)
	defer cancel()
	// ctx, cancel = context.WithTimeout(ctx, 50*time.Second)
	// defer cancel()
	if opt == singleTweet {
		handleTweet(ctx)
	}
}

func promptUser() (int, error) {
	var userOption int
	if DEBUG {
		userOption = singleTweet
		tweetData.auth_token = "c9bca772a8e05e076c17da20f126d22e042dae6b"
		tweetData.tweet_url = "https://x.com/rasjawa/status/1884869178123080001"
	}
	if !DEBUG {
		fmt.Println("=====CHOOSE MODE=====")
		fmt.Println("1. Search Mode")
		fmt.Println("2. Single Tweet Mode")
		fmt.Print("Your Option : ")
		fmt.Scan(&userOption)
	}

	if userOption == searchTweet {
		return 0, errors.New("Search mode is not yet implemented")
	} else if userOption == singleTweet {
		if !DEBUG {
			fmt.Print("Enter your twitter auth token : ")
			fmt.Scanln(&tweetData.auth_token)
			fmt.Print("Enter your desired twitter url : ")
			fmt.Scanln(&tweetData.tweet_url)
		}
		return singleTweet, nil
	} else {
		return 0, errors.New("Option provided is unknown")
	}
}

func scrollDown() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight})`, nil),
		chromedp.Evaluate(`document.querySelectorAll("a div[data-testid='tweetPhoto']").forEach((el) => el.remove())`, nil),
		chromedp.Evaluate(`document.evaluate("//span[contains(., 'Show replies')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext()?.click()`, nil),
		chromedp.Sleep(3 * time.Second),
	}
}
