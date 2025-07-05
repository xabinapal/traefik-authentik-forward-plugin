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
	if akPath == BasePath {
		return false
	}

	if strings.HasPrefix(akPath, BasePath+"/auth") {
		return false
	}

	return true
}

func GetAuthentikStartPath(u *url.URL) string {
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
