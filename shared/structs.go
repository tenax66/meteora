package shared

import "crypto/ed25519"

// the only struct for a meteora message
type Message struct {
	Id      string            `json:"id"`
	Content Content           `json:"content"`
	Pubkey  ed25519.PublicKey `json:"pubkey"`
	Sig     []byte            `json:"sig"`
}

type Content struct {
	Created_at int64  `json:"timestamp"`
	Text       string `json:"text"`
}
