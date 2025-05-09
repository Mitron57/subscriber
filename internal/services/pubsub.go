package services

import (
	"context"

	"github.com/mitron57/subpub"

	"subscriber/internal/domain/dto"
	"subscriber/internal/domain/interfaces"
)

type bus struct {
	inner subpub.SubPub
}

func NewPubSub(sp subpub.SubPub) interfaces.PubSub {
	return bus{
		inner: sp,
	}
}

func (b bus) Publish(ctx context.Context, message dto.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return b.inner.Publish(message.Topic, message.Data)
	}
}

func (b bus) Subscribe(ctx context.Context, subscription dto.Subscription) (subpub.Subscription, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return b.inner.Subscribe(subscription.Topic, subscription.Handler)
	}
}
