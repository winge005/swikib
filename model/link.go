package model

type Link struct {
	Id          int    `json:"id"`
	Category    string `json:"category"`
	Url         string `json:"url"`
	Description string `json:"description"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}
