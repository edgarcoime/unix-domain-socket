package pkg

import (
	"sync"
	"time"
)

var (
	counter int64
	mutex   sync.Mutex
	lastID  int64
)

func GenerateUniqueID() int64 {
	mutex.Lock()
	defer mutex.Unlock()

	// get current time in nano seconds
	currentTime := time.Now().UnixNano()

	// if the current time is the same as the last one increment otherwise reset
	if currentTime == lastID {
		counter++
	} else {
		counter = 0
	}

	// combine with counter for uniqueness
	uniqueID := currentTime + counter

	// update last id
	lastID = uniqueID

	return uniqueID
}
