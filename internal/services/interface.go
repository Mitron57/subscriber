package services

import (
	"context"

	"github.com/mitron57/subpub"

	"subscriber/internal/domain/dto"
)

type PubSub interface {
	// Publish does exactly what you think it does: delivers message to it's subject
	Publish(ctx context.Context, message dto.Message) error
	// Subscribe establishes new subscription to the target topic
	Subscribe(ctx context.Context, target dto.Subscription) (subpub.Subscription, error)
}
