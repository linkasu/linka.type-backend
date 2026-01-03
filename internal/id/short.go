package id

import (
	"math/rand"
	"time"
)

const (
	shortAlphabet = "abcdefghijklmnopqrstuvwxyz1234567890"
	shortSize     = 16
)

// NewShort returns a 16-char ID compatible with legacy Firebase functions.
func NewShort() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, shortSize)
	for i := 0; i < shortSize; i++ {
		buf[i] = shortAlphabet[rng.Intn(len(shortAlphabet))]
	}
	return string(buf)
}
