package tiktok

import (
	"bytes"
	"encoding/csv"
	"example/hello/pkg/terminal"
	"fmt"
	"os"
	"time"
)

var DEBUG = false

type Tiktok struct {
	TiktokQuery string
	TiktokResults   []TiktokScrapResult
}

var (
	SEARCH_MODE = "Search Tiktok Mode"
	SINGLE_MODE = "Single Tikotk Mode"
)

func (t *Tiktok) Setup()  {
    t.TiktokResults = make([]TiktokScrapResult, 0)
}

func (t *Tiktok) Begin()  {
	var mode string
	mode = SINGLE_MODE
	if !DEBUG {
		opt := terminal.Select{
			Opts:    []string{SEARCH_MODE, SINGLE_MODE},
			Message: "Choose mode",
		}
		var res int
		opt.Ask(&res)
		mode = opt.Opts[res]
	}
    t.Setup()
	if mode == SEARCH_MODE {
		panic("Search mode is not yet implemented")
	} else if mode == SINGLE_MODE {
        tiktokSingle := TiktokSingleOption{
            Tiktok: *t,
            HasMore: true,
        }
        tiktokSingle.Prompt()
        tiktokSingle.BeginSingleTiktok()
	} else {
        panic("Option provided is unknown")
	}
}


func (t *Tiktok) exportResultToCSV() (string, error) {
	fmt.Println("Starting the export process...")
	res := make([][]string, len(t.TiktokResults)+1)
	res[0] = []string{"Tiktok_ID", "Author", "Comment"}

	var i int = 1
	for _, val := range t.TiktokResults {
		res[i] = []string{val.TiktokId, val.Author, val.Content}
		i++
	}
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.WriteAll(res)
	if err := w.Error(); err != nil {
        return "", err
	}

	currentTime := time.Now().Local()
	fileName := fmt.Sprintf("res-tiktok%d.csv", currentTime.Unix())
	os.WriteFile(fileName, buf.Bytes(), 0644)
	fmt.Println("Successfully exported to : ", fileName)
	return fileName, nil
}
