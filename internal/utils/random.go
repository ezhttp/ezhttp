package utils

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/ezhttp/ezhttp/internal/logger"
)

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		logger.Error("error generating random bytes", "error", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func RandStringCharacters(count int) string {
	// Use crypto/rand for secure random generation
	b := make([]byte, count)

	// We need to avoid "modulo bias"
	// Our charset above (A-Z, a-z, 0-9) only has 62 characters
	// 62 does not evenly "fit" into a byte with value 0-255 (62 * 4 = 248 < 256)
	// So we use "rejection sampling" to chop off the extra range and
	// ensure each character has the exact same chance of appearing
	charsetLen := len(charset)
	maxValid := 256 - (256 % charsetLen)

	for i := 0; i < count; {
		randomBytes := make([]byte, count-i)
		_, err := rand.Read(randomBytes)
		if err != nil {
			logger.Error("error generating random string", "error", err)
			return ""
		}

		// Rejection sampling. If the byte value is greater
		// than our character set multiple, reject and regenerate
		for _, randomByte := range randomBytes {
			if int(randomByte) < maxValid {
				b[i] = charset[int(randomByte)%charsetLen]
				i++
				if i >= count {
					break
				}
			}
		}
	}

	return string(b)
}
