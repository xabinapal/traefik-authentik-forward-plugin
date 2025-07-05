package authentik

import (
	"net/http"
	"regexp"
)

type Config struct {
	Address                string
	CookiePolicy           http.SameSite
	UnauthorizedStatusCode int
	RedirectStatusCode     int
	UnauthorizedPaths      []*regexp.Regexp
	RedirectPaths          []*regexp.Regexp
}
