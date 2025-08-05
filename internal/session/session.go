package session

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
)

type Session struct {
	IsAuthenticated bool
	Headers         http.Header
	Cookies         []*http.Cookie
}

func GetIdentifier(cookies []*http.Cookie) string {
	if len(cookies) == 0 {
		return ""
	}

	// Collect key=value pairs
	pairs := make([]string, len(cookies))
	for i, c := range cookies {
		pairs[i] = c.Name + "=" + c.Value
	}

	sort.Strings(pairs)

	concat := ""
	for i, p := range pairs {
		if i > 0 {
			concat += ";"
		}
		concat += p
	}

	hash := sha256.Sum256([]byte(concat))
	return hex.EncodeToString(hash[:])
}