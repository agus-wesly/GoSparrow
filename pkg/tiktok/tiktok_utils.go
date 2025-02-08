package tiktok_pkg

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
