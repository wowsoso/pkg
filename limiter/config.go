package limiter

import (
	"time"
)

type Config struct {
	// Maximum number of requests to limit per duration.
	Max int `json:"max"`

	// Duration of rate-limiter.
	TTL time.Duration `json:"ttl"`

	Size int `json:"size"`
}
