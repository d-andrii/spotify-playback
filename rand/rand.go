package rand

import (
	"encoding/base64"
	"math/rand"
	"time"
)

func RandString(n int) string {
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	_, _ = rand.Read(b)

	return base64.StdEncoding.EncodeToString(b)
}
