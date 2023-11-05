package shared

type Message struct {
	Id      string  `json:"id"`
	Content Content `json:"content"`
}

type Content struct {
	Created_at int64  `json:"timestamp"`
	Text       string `json:"text"`
}
