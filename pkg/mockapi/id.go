package mockapi

import "math/rand/v2"

const lowercaseAlphanumericChars = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomLength(minLen, maxLen int) int {
	return rand.IntN(maxLen-minLen) + minLen //nolint:gosec
}

// ProjectID generates a random project ID.
func ProjectID() string {
	return lowercaseAlphanumericID(randomLength(10, 15))
}

// lowercaseAlphanumericID generates a random lowercase alphanumeric ID.
func lowercaseAlphanumericID(length int) string {
	id := make([]byte, length)
	for i := range id {
		id[i] = lowercaseAlphanumericChars[rand.IntN(len(lowercaseAlphanumericChars))] //nolint:gosec
	}

	return string(id)
}

// NumericID generates a random numeric ID.
func NumericID() string {
	length := randomLength(6, 10)
	id := make([]byte, length)
	for i := range id {
		id[i] = '0' + byte(rand.IntN(10)) //nolint:gosec
	}

	return string(id)
}
