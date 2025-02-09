package tiktok

type Response struct {
	Comments []Comment `json:"comments"`
    HasMore     int     `json:"has_more"`
    Cursor      int     `json:"cursor"`
}

type Comment struct {
	Text string `json:"text"`
	User User   `json:"user"`
}

type User struct {
    Uid     string  `json:"uid"`
	UniqueId string `json:"unique_id"`
    Nickname string `json:"nickname"`
}

type TiktokScrapResult struct {
	TiktokId  string `json:"tiktok_id"`
	Author    string `json:"username"`
	Content   string `json:"content"`
	UserIdStr string `json:"user_id_str"`
}

type TiktokData struct {
	TiktokUrl   string
	TiktokQuery string
}
