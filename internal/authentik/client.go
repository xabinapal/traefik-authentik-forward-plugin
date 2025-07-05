package authentik

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	client *http.Client
	config *Config
}

func NewClient(client *http.Client, config *Config) *Client {
	return &Client{
		client: client,
		config: config,
	}
}

func (c *Client) Check(meta *RequestMeta) (*ResponseMeta, error) {
	// send request to authentik to check if request is authenticated
	res, err := c.Request(meta, AuthPath, "")
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return &ResponseMeta{
			URL:             meta.URL,
			IsAuthenticated: false,
			Headers:         nil,
			Cookies:         GetCookies(res),
		}, nil
	case http.StatusOK:
		return &ResponseMeta{
			URL:             meta.URL,
			IsAuthenticated: true,
			Headers:         GetHeaders(res),
			Cookies:         GetCookies(res),
		}, nil
	default:
		return nil, fmt.Errorf("unexpected response: %d", res.StatusCode)
	}
}

func (c *Client) Request(meta *RequestMeta, path string, query string) (*http.Response, error) {
	// send request to authentik
	akReq, err := http.NewRequest(http.MethodGet, c.config.Address+path, nil)
	if err != nil {
		return nil, err
	}

	akReq.URL.RawQuery = query

	// add downstream request metadata
	akReq.Header.Set("X-Forwarded-Host", meta.URL.Host)
	akReq.Header.Set("X-Original-Uri", meta.URL.String())

	// add downstream authentik session cookies
	for _, c := range meta.Cookies {
		akReq.AddCookie(c)
	}

	res, err := c.client.Do(akReq)
	if err != nil {
		return nil, err
	}

	if err := c.mangleLocation(meta, res); err != nil {
		return nil, err
	}

	c.mangleCookies(res)

	return res, nil
}

func (c *Client) mangleLocation(meta *RequestMeta, res *http.Response) error {
	location := res.Header.Get("Location")
	if location == "" {
		res.Header.Del("Location")
		return nil
	}

	if strings.HasPrefix(location, c.config.Address+BasePath) {
		// convert absolute outpost redirects to downstream host
		location = strings.TrimPrefix(location, c.config.Address)

		locURL, err := url.Parse(location)
		if err != nil {
			return err
		}

		locURL.Scheme = meta.URL.Scheme
		locURL.Host = meta.URL.Host

		location = locURL.String()
	}

	res.Header.Set("Location", location)

	return nil
}

func (c *Client) mangleCookies(res *http.Response) {
	// get authentik session cookies from response
	cookies := GetCookies(res)

	res.Header.Del("Set-Cookie")
	for _, cookie := range cookies {
		res.Header.Add("Set-Cookie", cookie.String())
	}
}
