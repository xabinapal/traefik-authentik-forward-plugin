package authentik

import (
	"net/url"
	"strings"
)

const (
	BasePath  = "/outpost.goauthentik.io"
	StartPath = BasePath + "/start"
	AuthPath  = BasePath + "/auth"
	NginxPath = AuthPath + "/nginx"
)

func IsAuthentikPathAllowed(akPath string) bool {
	// allow all paths except the base path and auth paths
	return akPath != BasePath && !strings.HasPrefix(akPath, AuthPath)
}

func GetStartURL(u *url.URL) string {
	loc := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   StartPath,
		RawQuery: url.Values{
			"rd": {u.String()},
		}.Encode(),
	}

	return loc.String()
}
