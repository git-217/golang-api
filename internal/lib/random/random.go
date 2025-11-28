package random

import (
	"math/rand"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func NewRandomString(k int) string {
	b := make([]byte, k)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
