package authentik

import "strings"

const (
	BasePath = "/outpost.goauthentik.io"
	AuthPath = BasePath + "/auth/nginx"
)

func IsPathAllowedDownstream(akPath string) bool {
	if akPath == BasePath {
		return false
	}

	if strings.HasPrefix(akPath, BasePath+"/auth") {
		return false
	}

	return true
}
