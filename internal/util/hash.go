package util

import "crypto/md5"

func Hash(target string) (hashed string) {
	hasher := md5.New()
	hasher.Write([]byte(target))
	return string(hasher.Sum(nil))
}
