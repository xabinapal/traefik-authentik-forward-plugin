package authentik

import (
	"regexp"
)

const (
	BasePath = "/outpost.goauthentik.io"
	AuthPath = BasePath + "/auth/nginx"
)

var (
	AuthentikAllowedPaths = []*regexp.Regexp{
		regexp.MustCompile("^" + BasePath + "/auth/start/?" + "$"),
	}

	AuthentikRestrictedPaths = []*regexp.Regexp{
		regexp.MustCompile("^" + BasePath + "/?" + "$"),
		regexp.MustCompile("^" + BasePath + "/auth/.*" + "$"),
	}
)

func IsPathAllowed(akPath string) bool {
	for _, pattern := range AuthentikAllowedPaths {
		if pattern.MatchString(akPath) {
			return true
		}
	}

	for _, pattern := range AuthentikRestrictedPaths {
		if pattern.MatchString(akPath) {
			return false
		}
	}

	return true
}
