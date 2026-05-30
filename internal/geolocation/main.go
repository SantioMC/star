package geolocation

import (
	"context"
)

type Service interface {
	GetLocation(ctx context.Context, ip string) (*string, error)
}

func isLocal(ip string) bool {
	return ip == "127.0.0.1" || ip == "[::1]"
}
