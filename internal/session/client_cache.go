package session

import (
	"context"
	"sync"
	"time"
)

type CacheClient struct {
	context  context.Context
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

func (c *CacheClient) Get(session string) *Session {
	if session, ok := c.store.Load(session); ok {
		return session.(*Session)
	}

	return nil
}

func (c *CacheClient) Set(session string, meta *Session) {
	c.store.Store(session, meta)
	if c.context.Err() != nil {
		return
	}

	go func() {
		timer := time.NewTimer(c.duration)
		defer timer.Stop()

		select {
		case <-timer.C:
			if c.context.Err() == nil {
				c.store.Delete(session)
			}
		case <-c.context.Done():
		}
	}()
}

func (c *CacheClient) Delete(session string) {
	c.store.Delete(session)
}
