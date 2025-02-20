package main

import (
	"example/hello/pkg/terminal"
	"example/hello/pkg/twitter"
	"log"
)

const (
	TWITTER = "Twitter"
	TIKTOK  = "Tiktok"
)

var foo []string

var DEBUG bool = false

var Test struct {
	Foo string
}

func main() {
    var resp string
    inp := terminal.Input{Message: "Enter your tweet search keyword : ", Required: true}; 
    err := inp.Ask(&resp)
    if err != nil {
        log.Fatalln(err)
    }
    if true {
        return
    }
	prompt := terminal.Select{
		Opts:    []string{TWITTER, TIKTOK},
		Message: "Choose Social Media Mode",
	}
	var res int
	err = prompt.Ask(&res)
	if err != nil {
		panic(err)
	}
	selectedSocialMedia := prompt.Opts[res]

	if selectedSocialMedia == TWITTER {
		t := twitter.Tweet{}
		t.Begin()
	} else if selectedSocialMedia == TIKTOK {
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
