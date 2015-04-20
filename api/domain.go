package api

import (
	"time"
)

type Content struct {
	UUID            string    `json:"uuid"`
	LastPublishDate time.Time `json:"last-published-time"`
	Headline        string    `json:"title"`
	TextBody        string    `json:"bodyPlain"`
	Link            string    `json:"link,omitempty"`
	Genre           string    `json:"genre,omitempty"`
}
