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
	BodyXML       string    `json:"bodyXML,omitempty"`
	Brands        []string  `json:"brands,omitempty"`
	Byline        string    `json:"byline,omitempty"`
	PublishedDate time.Time `json:"publishedDate,omitempty"`
	//RequestUrl    string   `json:"requestUrl"`
	Title  string `json:"title,omitempty"`
	//Type   string `json:"type"`
	WebUrl string `json:"webUrl,omitempty"`
}
