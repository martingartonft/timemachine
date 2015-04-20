package api

import (
	"errors"
)

type ContentAPI interface {
	Count() int
	ByUUID(id string) (bool, Content)
	Write(c Content) error
	All(stopchan chan struct{}) (chan Content, error)
	Recent(stopChan chan struct{}, limit int) (chan Content, error)
	Drop() error
	Close()
}

var (
	ERR_NOT_IMPLEMENTED = errors.New("not implemented")
	ERR_INVALID_QUERY   = errors.New("invalid query")
)