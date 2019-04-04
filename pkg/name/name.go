package name

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func Limit(s string, count int) string {
	if len(s) < count {
		return s
	}
	return fmt.Sprintf("%s-%s", s[:count-6], Hex(s, 5))
}

func Hex(s string, length int) string {
	h := md5.Sum([]byte(s))
	d := hex.EncodeToString(h[:])
	return d[:length]
}

func PublicDomain(s string) string {
	return Limit(strings.Replace(s, ".", "-", -1), 15)
}

func SafeConcatName(name ...string) string {
	fullPath := strings.Join(name, "-")
	if len(fullPath) > 63 {
		digest := sha256.Sum256([]byte(fullPath))
		return fullPath[0:57] + "-" + string(digest[:])[0:5]
	}
	return fullPath
}
