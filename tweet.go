package main

import (
	"errors"
	"fmt"
)

const (
	searchTweet = iota + 1
	singleTweet
)

func promptTweet() (int, error) {
	var userOption int
	if DEBUG {
		userOption = singleTweet
		tweetData.auth_token = "c9bca772a8e05e076c17da20f126d22e042dae6b"
		tweetData.tweet_url = "https://x.com/investorgabut/status/1885126576725237962"
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
