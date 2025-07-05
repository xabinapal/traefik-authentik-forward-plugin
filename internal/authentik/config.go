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
	// check if request path matches any of the skipped paths
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

	// check if request path matches any of the unauthorized paths
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

	// check if request path matches any of the redirect paths
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
		// return the status code of the longest match
		return longestMatchStatusCode
	}

	// allow request if no match is found
	return http.StatusOK
}
