package main

import (
	"example/hello/pkg/terminal"
	"example/hello/pkg/tiktok"
	"example/hello/pkg/twitter"
)

const (
	TWITTER = "Twitter"
	TIKTOK  = "Tiktok"
)

var foo []string

var DEBUG bool = false

func main() {
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
