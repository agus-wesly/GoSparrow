package twitter

type Tweet struct {
	AuthToken  string
	TweetUrl   string
	TweetQuery string
}

type Response struct {
	Data ResponseData `json:"data"`
}

type ResponseData struct {
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
	FullText  string `json:"full_text"`
	UserIdStr string `json:"user_id_str"`
}

type TweetScrapResult struct {
	TweetId   string `json:"tweet_id"`
	Author    string `json:"author_id"`
	Content   string `json:"content"`
	UserIdStr string `json:"user_id_str"`
}
