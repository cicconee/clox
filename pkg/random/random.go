package random

import "math/rand"

func ID(l int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	state := make([]byte, l)
	for i := range state {
		state[i] = charset[rand.Intn(len(charset))]
	}

	return string(state)
}
