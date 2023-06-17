package main

import (
	"github.com/google/uuid"
	"sync"
)

func getLocks(c configuration) map[uuid.UUID]*sync.Mutex {
	locks := make(map[uuid.UUID]*sync.Mutex)
	for _, ascript := range c.Scripts {
		if !ascript.Concurrent {
			locks[ascript.ID] = new(sync.Mutex)
		}
	}
	return locks
}
