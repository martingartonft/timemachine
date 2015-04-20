package api

import (
	"time"
)

/*
type Content struct {
	UUID            string    `json:"uuid"`
	LastPublishDate time.Time `json:"last-published-time"`
	Headline        string    `json:"title"`
	TextBody        string    `json:"bodyPlain"`
	Link            string    `json:"link,omitempty"`
	Genre           string    `json:"genre,omitempty"`
}
*/
type Content struct {
	//ID            string   `json:"id"`
	UUID          string    `json:"uuid"`
	BodyXML       string    `json:"bodyXML"`
	Brands        []string  `json:"brands"`
	Byline        string    `json:"byline"`
	PublishedDate time.Time `json:"publishedDate"`
	//RequestUrl    string   `json:"requestUrl"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	WebUrl string `json:"webUrl"`
}
