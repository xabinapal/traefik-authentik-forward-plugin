package session

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type CacheClient struct {
	context  context.Context //nolint:containedctx
	duration time.Duration
	store    sync.Map
}

func NewCacheClient(context context.Context, duration time.Duration) *CacheClient {
	if duration == 0 {
		panic("duration must be greater than 0")
	}

	return &CacheClient{
		context:  context,
		duration: duration,
		store:    sync.Map{},
	}
}

func (c *CacheClient) Get(cookies []*http.Cookie) *Session {
	sessionId := GetIdentifier(cookies)
	if v, ok := c.store.Load(sessionId); ok {
		if s, ok := v.(*Session); ok {
			return s
		}
	}

	return nil
}

func (c *CacheClient) Set(cookies []*http.Cookie, meta *Session) {
	sessionId := GetIdentifier(cookies)
	c.store.Store(sessionId, meta)
	if c.context.Err() != nil {
		return
	}

	go func() {
		timer := time.NewTimer(c.duration)
		defer timer.Stop()

		select {
		case <-timer.C:
			if c.context.Err() == nil {
				c.store.Delete(sessionId)
			}
		case <-c.context.Done():
		}
	}()
}

func (c *CacheClient) Delete(cookies []*http.Cookie) {
	sessionId := GetIdentifier(cookies)
	c.store.Delete(sessionId)
}
