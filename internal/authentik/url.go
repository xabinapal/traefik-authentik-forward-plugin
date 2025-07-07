package authentik

import (
	"net/url"
	"strings"
)

const (
	BasePath = "/outpost.goauthentik.io"
	AuthPath = BasePath + "/auth/nginx"
)

func IsAuthentikPathAllowed(akPath string) bool {
	// allow all paths except the base path and auth paths
	return akPath != BasePath && !strings.HasPrefix(akPath, BasePath+"/auth")
}

func GetStartURL(u *url.URL) string {
	loc := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   BasePath + "/start",
		RawQuery: url.Values{
			"rd": {u.String()},
		}.Encode(),
	}

	return loc.String()
}
