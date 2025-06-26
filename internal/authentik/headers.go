package authentik

import (
	"net/http"
	"strings"
)

const (
	HeaderPrefix = "X-Authentik-"
)

func GetHeaders(res *http.Response) http.Header {
	headers := http.Header{}
	for k, v := range res.Header {
		if strings.HasPrefix(k, HeaderPrefix) && k != HeaderPrefix {
			headers[k] = v
		}
	}

	return headers
}
