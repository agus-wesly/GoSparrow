package twitter

type Response struct {
	Data ResponseData `json:"data"`
}

type ResponseData struct {
	ThreadedConversationsWithInjectionsV2 ThreadedConversation `json:"threaded_conversation_with_injections_v2"`
}

type ThreadedConversation struct {
	Instructions []Instruction `json:"instructions"`
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
	RestId string           `json:"rest_id"`
	Legacy UserResultLegacy `json:"legacy"`
}

type UserResultLegacy struct {
	Name string `json:"name"`
}

//

type Legacy struct {
	FullText  string `json:"full_text"`
	UserIdStr string `json:"user_id_str"`
}

type TweetScrapResult struct {
	TweetId   string `json:"tweet_id"`
	Author    string `json:"author_id"`
	Content   string `json:"content"`
	UserIdStr string `json:"user_id_str"`
}

type SearchResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	SearchByRawQuery SearchByRawQuery `json:"search_by_raw_query"`
}

type SearchByRawQuery struct {
	SearchTimeline SearchTimeline `json:"search_timeline"`
}

type SearchTimeline struct {
	Timeline Timeline `json:"timeline"`
}

type Timeline struct {
	Instructions []Instruction `json:"instructions"`
}

// tweets = responseJson.data?.search_by_raw_query.search_timeline.timeline?.instructions?.[0]?.entries;
type Instruction struct {
	Entries []Entry `json:"entries"`
}
