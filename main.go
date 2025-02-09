package main

import (
	"context"
	"errors"
	"example/hello/pkg/twitter"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

const (
	twitterOption = iota + 1
	tiktokOption
)

var DEBUG bool = false
var tweetData twitter.Tweet

func basePrompt() (int, error) {
	var userOption int
	fmt.Println("=====CHOOSE SOCIAL MEDIA MODE=====")
	fmt.Println("1. Twitter")
	fmt.Println("2. Tiktok")
	fmt.Print("Your Option : ")
	fmt.Scan(&userOption)

	if userOption == twitterOption || userOption == tiktokOption {
		return userOption, nil
	}

	return 0, errors.New("Bad option")
}

func createNewContext() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
	)
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(actx)
	return ctx, acancel
}

func main() {

	opt, err := basePrompt()
	if err != nil {
		log.Fatalln(err)
	}

	// Twitter
	if opt == twitterOption {
		opt_tweet, err_tweet := promptTweet()
		if err_tweet != nil {
			log.Fatalln(err)
		}

		if opt_tweet == singleTweet {
			ctx, acancel := createNewContext()
			handleSingleTweet(ctx)
			defer acancel()
		}
		// Tiktok
	} else if opt == tiktokOption {
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
