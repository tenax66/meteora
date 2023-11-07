package shared

// the only struct for a meteora message
type Message struct {
	Id      string  `json:"id"`
	Content Content `json:"content"`
	// Pubkey
	// Sig
}

type Content struct {
	Created_at int64  `json:"timestamp"`
	Text       string `json:"text"`
}
