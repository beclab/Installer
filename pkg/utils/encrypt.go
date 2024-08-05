package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5(str string) string {
	hasher := md5.New()

	hasher.Write([]byte(str))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}
