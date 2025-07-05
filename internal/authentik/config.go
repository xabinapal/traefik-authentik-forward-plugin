package authentik

import (
	"regexp"
)

type Config struct {
	Address                string
	UnauthorizedStatusCode int
	RedirectStatusCode     int
	UnauthorizedPaths      []*regexp.Regexp
	RedirectPaths          []*regexp.Regexp
}
