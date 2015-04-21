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
	URI           string    `json:"id"`
	BodyXML       string    `json:"bodyXML,omitempty"`
	Brands        []string  `json:"brands,omitempty"`
	Byline        string    `json:"byline,omitempty"`
	PublishedDate time.Time `json:"publishedDate,omitempty"`
	//RequestUrl    string   `json:"requestUrl"`
	Title string `json:"title,omitempty"`
	//Type   string `json:"type"`
	WebUrl string `json:"webUrl,omitempty"`
}

type Version struct {
	UUID          string    `json:"uuid"`
	Version       string    `json:"version"`
	PublishedDate time.Time `json:"publishedDate,omitempty"`
}

type Versions []Version

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Less(a int, b int) bool {
	da := v[a].PublishedDate
	db := v[b].PublishedDate
	return da.Before(db)
}

func (v Versions) Swap(a int, b int) {
	v[a], v[b] = v[b], v[a]
}
