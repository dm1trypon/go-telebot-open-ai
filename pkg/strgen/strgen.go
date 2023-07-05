package strgen

import (
	"math/rand"
	"strings"
	"time"
)

var dictionary = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func Generate(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(dictionary[r.Intn(len(dictionary))])
	}
	return b.String()
}
