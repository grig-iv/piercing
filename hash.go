package main

import (
	"crypto/sha256"
)

func getHash(str string) []byte {
	h := sha256.New()

	h.Write([]byte(str))

	return h.Sum(nil)
}
