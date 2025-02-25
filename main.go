package main

import (
	"example/hello/pkg/terminal"
	"example/hello/pkg/tiktok"
	"example/hello/pkg/twitter"
	"fmt"

	"github.com/mgutz/ansi"
)

const (
	TWITTER = "Twitter"
	TIKTOK  = "Tiktok"
)

var foo []string

var DEBUG bool = false

func main() {
    promptHeader()
	prompt := terminal.Select{
		Opts:    []string{TWITTER, TIKTOK},
		Message: "Choose Social Media Mode",
	}
	var res int
	err := prompt.Ask(&res)
	if err != nil {
		panic(err)
	}
	selectedSocialMedia := prompt.Opts[res]

	if selectedSocialMedia == TWITTER {
		t := twitter.Tweet{}
		t.Begin()
	} else if selectedSocialMedia == TIKTOK {
		t := tiktok.Tiktok{}
		t.Begin()
	}
}

func promptHeader() {
	fmt.Println(ansi.LightCyan)
	fmt.Println(` _________         ________                                             `)
	fmt.Println(` __  ____/_____    __  ___/_____________ _______________________      __`)
	fmt.Println(` _  / __ _  __ \   _____ \___  __ \  __ ` + "`" + `/_  ___/_  ___/  __ \_ | /| / /`)
	fmt.Println(` / /_/ / / /_/ /   ____/ /__  /_/ / /_/ /_  /   _  /   / /_/ /_ |/ |/ / `)
	fmt.Println(` \____/  \____/    /____/ _  .___/\__,_/ /_/    /_/    \____/____/|__/  `)
	fmt.Println(`                            /_/                                           `)
	fmt.Println(`                                            Created with 🤍 by : Wesly  `)
	fmt.Println(ansi.Reset)
}
