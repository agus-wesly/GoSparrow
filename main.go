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
	twitter = iota + 1
	tiktok
)

var tweetData TweetData

var DEBUG bool = true

func prompt() (int, error) {
	var userOption int
	fmt.Println("=====CHOOSE SOCIAL MEDIA MODE=====")
	fmt.Println("1. Twitter")
	fmt.Println("2. Tiktok")
	fmt.Print("Your Option : ")
	fmt.Scan(&userOption)

	if userOption == twitter {
		return twitter, nil
	}

	if userOption == tiktok {
		return tiktok, nil
	}

	return 0, errors.New("Bad option")
}

func createNewContext() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(actx)
	return ctx, acancel
}

func main() {

	opt, err := prompt()
	if err != nil {
		log.Fatalln(err)
	}

	if opt == twitter {
		opt_tweet, err_tweet := promptTweet()
		if err_tweet != nil {
			log.Fatalln(err)
		}

		if opt_tweet == singleTweet {
			ctx, acancel := createNewContext()
			handleSingleTweet(ctx)
			defer acancel()
		}
	} else if opt == tiktok {
		opt_tiktok, err_tiktok := promptTiktok()
		if err_tiktok != nil {
			log.Fatalln(err_tiktok)
		}
		if opt_tiktok == singleTiktok {
			handleSingleTiktok()
		}
	}

	// ctx, cancel = context.WithTimeout(ctx, 50*time.Second)
	// defer cancel()
}

func scrollDown() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight})`, nil),
		chromedp.Evaluate(`document.querySelectorAll("a div[data-testid='tweetPhoto']").forEach((el) => el.remove())`, nil),
		chromedp.Evaluate(`document.evaluate("//span[contains(., 'Show replies')]", document, null, XPathResult.ANY_TYPE, null ).iterateNext()?.click()`, nil),
		chromedp.Sleep(3 * time.Second),
	}
}
