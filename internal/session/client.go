package session

import (
	"context"
	"net/http"
	"time"
)

type Client interface {
	Get(cookies []*http.Cookie) *Session
	Set(cookies []*http.Cookie, meta *Session)
	Delete(cookies []*http.Cookie)
}

func NewClient(context context.Context, duration time.Duration) Client { //nolint:ireturn
	if duration == 0 {
		return NewStandardClient()
	}

	return NewCacheClient(context, duration)
}
