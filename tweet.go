package main

import (
	"bufio"
	"example/hello/pkg/twitter"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chromedp/chromedp"
)

const (
	searchTweet = iota + 1
	singleTweet
)

func promptTweet() {
	var userOption int
	if DEBUG {
        userOption = searchTweet
		tweetData.AuthToken = "c9bca772a8e05e076c17da20f126d22e042dae6b"
	} else {
		fmt.Print("Enter your twitter auth token : ")
		fmt.Scanln(&tweetData.AuthToken)
	}
	ctx, acancel := createNewContext()
    err := chromedp.Run(
        ctx,
		authenticateTwitter("auth_token", tweetData.AuthToken),
	    verifyLoginTwitter(),
    )
    if err != nil {
        fmt.Println("Auth token is invalid", err)
        os.Exit(1)
    }
    acancel()
	if !DEBUG {
		fmt.Println("=====CHOOSE MODE=====")
		fmt.Println("1. Search Mode")
		fmt.Println("2. Single Tweet Mode")
		fmt.Print("Your Option : ")
		fmt.Scan(&userOption)
	}
	if (userOption != 1) && (userOption != 2) {
		log.Fatalln("Unknown option")
	}

	if userOption == singleTweet {
		var twitterSingleOption twitter.SingleOption
        twitterSingleOption.TweetUrl = "https://x.com/ctjlewis/status/1890113604806164798"
		if !DEBUG {
			fmt.Print("Enter your desired twitter url : ")
			fmt.Scanln(&twitterSingleOption.TweetUrl)
		}
		beginSingleTweet(twitterSingleOption.TweetUrl)
	} else if userOption == searchTweet {
		twitterSearchOption := twitter.SearchOption{
			MinReplies: 0,
			Query:      "",
			MinLikes:   0,
			Language:   "en",
		}
		if !DEBUG {
			fmt.Print("Enter your tweet search keyword : ")
			in := bufio.NewReader(os.Stdin)
			inputQuery, err := in.ReadString('\n')
			if err != nil {
				log.Fatalln(err)
			}
			inputQuery = strings.TrimSpace(inputQuery)
			twitterSearchOption.Query = inputQuery
			fmt.Print("Minimum tweet replies (Default=0) : ")
			fmt.Scan(&twitterSearchOption.MinReplies)
			fmt.Print("Minimum tweet likes (Default=0) : ")
			fmt.Scan(&twitterSearchOption.MinLikes)
			fmt.Print("Language [en/id] (Default=en) : ")
			fmt.Scan(&twitterSearchOption.Language)
		} else {
			twitterSearchOption.Query = "var indonesia"
			twitterSearchOption.MinReplies = 10
		}
        beginSearchTweet(&twitterSearchOption)
	}
}
