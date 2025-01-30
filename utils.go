package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	ThreadedConversationsWithInjectionsV2 ThreadedConversation `json:"threaded_conversation_with_injections_v2"`
}

type ThreadedConversation struct {
	Instructions []Instruction `json:"instructions"`
}

type Instruction struct {
	Type    string  `json:"type"`
	Entries []Entry `json:"entries"`
}

type Entry struct {
	EntryId   string  `json:"entryId"`
	SortIndex string  `json:"sortIndex"`
	Content   Content `json:"content"`
}

type Content struct {
	EntryType   string       `json:"entryType"`
	ItemContent *ItemContent `json:"itemContent"`
	Items       *[]Items     `json:"items"`
}

type Items struct {
	EntryId string `json:"entryId"`
	Item    Item   `json:"item"`
}

type Item struct {
	ItemType     string       `json:"itemType"`
	TweetResults TweetResults `json:"tweet_results"`
	ItemContent  *ItemContent `json:"itemContent"`
}

type ItemContent struct {
	ItemType     string       `json:"itemType"`
	TweetResults TweetResults `json:"tweet_results"`
}

type TweetResults struct {
	Result Result `json:"result"`
}

type Result struct {
	RestId string `json:"rest_id"`
	Legacy Legacy `json:"legacy"`
	Core   Core   `json:"core"`
}

// Username //
type Core struct {
	Results UserResults `json:"user_results"`
}

type UserResults struct {
	Result UserResult `json:"result"`
}

type UserResult struct {
	Legacy UserResultLegacy `json:"legacy"`
}

type UserResultLegacy struct {
	Name string `json:"name"`
}

//

type Legacy struct {
	FullText string `json:"full_text"`
}

type TweetScrapResult struct {
	TweetId string `json:"tweet_id"`
	Author  string `json:"author_id"`
	Content string `json:"content"`
}

func processTweetJSON(jsonData Response) {
	var entries []Entry
	entries = jsonData.Data.ThreadedConversationsWithInjectionsV2.Instructions[0].Entries
	for i := 0; i < len(entries); i++ {
		currentEntryContent := entries[i].Content
		var item *ItemContent
		if currentEntryContent.ItemContent != nil {
			item = currentEntryContent.ItemContent
			addToTweetResult(item)
		}

		if currentEntryContent.Items != nil {
			items := *currentEntryContent.Items
			for j := 0; j < len(items); j++ {
				item = items[j].Item.ItemContent
				addToTweetResult(item)
			}
		}
	}
}

func saveToJsonFile(data []byte) {
	err := os.WriteFile("tweet-response.json", data, 0644)
	if err != nil {
		log.Fatalln("error saving to json:", err)
	}
}

func addToTweetResult(item *ItemContent) {
	tweet := TweetScrapResult{
		TweetId: item.TweetResults.Result.RestId,
		Author:  item.TweetResults.Result.Core.Results.Result.Legacy.Name,
		Content: item.TweetResults.Result.Legacy.FullText,
	}
	tweet_results[tweet.TweetId] = tweet
}

// var tweet_results = make(map[string]TweetScrapResult)
// id,col1,col2
// id_1,340.384926,123.285031
// id_1,321.385028,4087.284675
func exportToCSV() {
	res := make([][]string, len(tweet_results)+1)
	res[0] = []string{"Tweet_Id", "Author", "Content"}

	var i int = 1
	for _, val := range tweet_results {
		res[i] = []string{val.TweetId, val.Author, val.Content}
		i++
	}
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.WriteAll(res)

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
	currentTime := time.Now().Local()
	os.WriteFile(fmt.Sprintf("res-%d.csv", currentTime.Unix()), buf.Bytes(), 0644)
}
