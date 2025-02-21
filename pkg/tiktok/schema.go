package tiktok

type ResponseJson struct {
	Comments []Comment `json:"comments"`
	HasMore  int       `json:"has_more"`
	Cursor   int       `json:"cursor"`
}

type Comment struct {
	Text string `json:"text"`
	User User   `json:"user"`
}

type User struct {
	Uid      string `json:"uid"`
	UniqueId string `json:"unique_id"`
	Nickname string `json:"nickname"`
}

type TiktokScrapResult struct {
	TiktokId  string `json:"tiktok_id"`
	Author    string `json:"username"`
	Content   string `json:"content"`
	UserIdStr string `json:"user_id_str"`
}

type VideoListResponse struct {
	Data []Response `json:"data"`
}

type Response struct {
	Item Item `json:"item"`
}

type Item struct {
	Video  Video  `json:"video"`
	Author Author `json:"author"`
}

type Video struct {
	Id string `json:"id"`
}

type Author struct {
	Id       string `json:"string"`
	UniqueId string `json:"uniqueId"`
}
