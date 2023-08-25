package ioc

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
)

func InitMemory() *bigcache.BigCache {
	cache, err := bigcache.New(context.Background(), bigcache.Config{
		Shards:             256,
		LifeWindow:         10 * time.Minute,
		CleanWindow:        10 * time.Minute,
		MaxEntriesInWindow: 1024,
		MaxEntrySize:       64,
		Verbose:            true,
	})
	if err != nil {
		panic(err)
	}

	return cache
}
