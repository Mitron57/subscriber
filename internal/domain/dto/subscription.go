package dto

import "github.com/mitron57/subpub"

type Subscription struct {
	Topic   string
	Handler subpub.MessageHandler
}
