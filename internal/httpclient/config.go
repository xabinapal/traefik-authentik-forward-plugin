package httpclient

import (
	"time"
)

type Config struct {
	Timeout time.Duration
	TLS     TLSConfig
}

type TLSConfig struct {
	CA                 string
	Cert               string
	Key                string
	MinVersion         uint16
	MaxVersion         uint16
	InsecureSkipVerify bool
}
