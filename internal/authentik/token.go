package authentik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/xabinapal/traefik-authentik-forward-plugin/internal/session"
)

const (
	// TokenUserPath is the Authentik API endpoint used to validate a bearer
	// token and retrieve the user it belongs to. The /me/ endpoint wraps the
	// payload under a "user" key.
	TokenUserPath = "/api/v3/core/users/me/"

	// GroupsSeparator matches the separator Authentik outposts use to join
	// group names in the X-Authentik-Groups header.
	GroupsSeparator = "|"

	// tokenCacheCookieName is a synthetic cookie name used to derive a session
	// cache key from a bearer token, reusing the existing session cache.
	tokenCacheCookieName = "authentik_token"
)

// tokenUser mirrors the relevant fields returned by /api/v3/core/users/me/.
type tokenUser struct {
	User struct {
		PK       json.Number `json:"pk"`
		Username string      `json:"username"`
		Email    string      `json:"email"`
		Groups   []struct {
			Name string `json:"name"`
		} `json:"groups"`
	} `json:"user"`
}

// tokenCacheKey derives a deterministic session cache identifier for a bearer
// token by reusing the cookie-based session cache. The token value is hashed by
// the session store, so it is never stored in plaintext.
func tokenCacheKey(token string) []*http.Cookie {
	return []*http.Cookie{{Name: tokenCacheCookieName, Value: token}}
}

// checkToken validates a bearer token against Authentik, caching the result so
// repeated requests with the same token don't overload the server.
func (c *Client) checkToken(meta *RequestMeta) (*ResponseMeta, error) {
	cacheKey := tokenCacheKey(meta.Token)

	// check if the token validation is already cached
	if s := c.session.Get(cacheKey); s != nil {
		return &ResponseMeta{
			URL:     meta.URL,
			Cached:  true,
			Session: s,
		}, nil
	}

	s, err := c.requestToken(meta.Token)
	if err != nil {
		return nil, err
	}

	// cache token validation result
	c.session.Set(cacheKey, s)

	return &ResponseMeta{
		URL:     meta.URL,
		Cached:  false,
		Session: s,
	}, nil
}

// requestToken calls Authentik to validate the bearer token and builds a
// session with the same X-Authentik-* headers the outpost would set.
func (c *Client) requestToken(token string) (*session.Session, error) {
	akReq, err := http.NewRequest(http.MethodGet, c.config.Address+TokenUserPath, nil)
	if err != nil {
		return nil, err
	}

	akReq.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(akReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	// any non-2xx response means the token is invalid or expired, mirroring the
	// reference implementation that treats RestClientResponseException as null
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return &session.Session{IsAuthenticated: false}, nil
	}

	var body tokenUser
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	user := body.User
	if user.PK.String() == "" || user.Username == "" {
		return &session.Session{IsAuthenticated: false}, nil
	}

	groups := make([]string, 0, len(user.Groups))
	for _, g := range user.Groups {
		if g.Name != "" {
			groups = append(groups, g.Name)
		}
	}

	headers := http.Header{}
	headers.Set(HeaderPrefix+"Uid", user.PK.String())
	headers.Set(HeaderPrefix+"Username", user.Username)
	headers.Set(HeaderPrefix+"Email", user.Email)
	headers.Set(HeaderPrefix+"Groups", strings.Join(groups, GroupsSeparator))

	return &session.Session{
		IsAuthenticated: true,
		Headers:         headers,
		Cookies:         nil,
	}, nil
}
