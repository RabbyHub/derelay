package relay

import (
	"crypto/rand"
	"encoding/base64"
)

func generateRandomBytes16() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf)
}
