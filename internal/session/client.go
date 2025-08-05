package session

import (
	"context"
	"time"
)

type Client interface {
	Get(session string) *Session
	Set(session string, meta *Session)
	Delete(session string)
}

func NewClient(context context.Context, duration time.Duration) Client {
	if duration == 0 {
		return NewStandardClient()
	}

	return NewCacheClient(context, duration)
}
