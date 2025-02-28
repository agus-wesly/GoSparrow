package instagram

type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	Connection CommentsConnection `json:"xdt_api__v1__media__media_id__comments__connection"`
}

type CommentsConnection struct {
	Edges []interface{} `json:"edges"`
}
