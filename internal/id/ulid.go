package id

import (
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	mu      sync.Mutex
)

// New returns a monotonic ULID string.
func New() string {
	mu.Lock()
	defer mu.Unlock()
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}
