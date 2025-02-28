package main

import (
	"flag"
	"fmt"

	"github.com/agus-wesly/GoSparrow/pkg/core"
	"github.com/agus-wesly/GoSparrow/pkg/instagram"
	"github.com/agus-wesly/GoSparrow/pkg/terminal"
	"github.com/agus-wesly/GoSparrow/pkg/tiktok"
	"github.com/agus-wesly/GoSparrow/pkg/twitter"

	"github.com/mgutz/ansi"
)

const (
	TWITTER = "Twitter"
	TIKTOK  = "Tiktok"
	INSTAGRAM  = "Instagram"
)

var foo []string

var DEBUG bool = true

func main() {
	headless := flag.Bool("headless", false, "Specify if app run in the headless mode")
	flag.Parse()
	core.IS_HEADLESS = *headless

    ins := instagram.Instagram{}
    ins.Begin()

    if true {
        return
    }
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
	} else if selectedSocialMedia == INSTAGRAM {
        //ins := instagram.Instagram{}
        //ins.Begin()
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
	fmt.Println(`                                            Scrap any social media  `)
	fmt.Println(`                                            Created with 🤍 by : Wesly:)  `)
	fmt.Println(ansi.Reset)
}
