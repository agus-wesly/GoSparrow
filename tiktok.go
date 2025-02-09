package main

import (
	"errors"
	"example/hello/pkg/tiktok"
	"fmt"
)

const (
	searchTiktok = iota + 1
	singleTiktok
)


var tiktokData tiktok.TiktokData


func promptTiktok() (int, error) {
	var userOption int
	if DEBUG {
		userOption = singleTiktok
		tiktokData.TiktokUrl = "https://www.tiktok.com/@andwey.kurt1/video/7467314067489721608"
	}
	if !DEBUG {
		fmt.Println("=====CHOOSE MODE=====")
		fmt.Println("1. Search Mode")
		fmt.Println("2. Single tiktok Mode")
		fmt.Print("Your Option : ")
		fmt.Scan(&userOption)
	}

	if userOption == searchTiktok {
		return 0, errors.New("Search mode is not yet implemented")
	} else if userOption == singleTiktok {
		if !DEBUG {
			fmt.Print("Enter your desired tiktok url : ")
			fmt.Scanln(&tiktokData.TiktokUrl)
		}
		return singleTiktok, nil
	} else {
		return 0, errors.New("Option provided is unknown")
	}
}
