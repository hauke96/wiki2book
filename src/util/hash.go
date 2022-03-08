package util

import (
	"crypto/sha1"
	"encoding/hex"
)

func Hash(data string) string {
	hash := sha1.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}
