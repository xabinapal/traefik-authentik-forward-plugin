package authentik

import (
	"net/http"
	"regexp"
)

type Config struct {
	Address string

	UnauthorizedStatusCode int
	RedirectStatusCode     int

	SkippedPaths      []*regexp.Regexp
	UnauthorizedPaths []*regexp.Regexp
	RedirectPaths     []*regexp.Regexp
}

func (c *Config) IsSkippedPath(path string) bool {
	for _, p := range c.SkippedPaths {
		if p.MatchString(path) {
			return true
		}
	}

	return false
}

func (c *Config) GetUnauthorizedStatusCode(path string) int {
	var longestMatch *regexp.Regexp
	var longestMatchLength int
	var longestMatchStatusCode int

	for _, re := range c.UnauthorizedPaths {
		if re.MatchString(path) {
			l := len(re.String())
			if l > longestMatchLength {
				longestMatch = re
				longestMatchLength = l
				longestMatchStatusCode = c.UnauthorizedStatusCode
			}
		}
	}

	for _, re := range c.RedirectPaths {
		if re.MatchString(path) {
			l := len(re.String())
			if l > longestMatchLength {
				longestMatch = re
				longestMatchLength = l
				longestMatchStatusCode = c.RedirectStatusCode
			}
		}
	}

	if longestMatch != nil {
		return longestMatchStatusCode
	}

	return http.StatusOK
}
