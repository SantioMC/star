package geolocation

import (
	"context"
	"errors"
)

var stubLocalLocation = "Localhost"

type StubGeolocationService struct{}

func (s *StubGeolocationService) GetLocation(ctx context.Context, ip string) (*string, error) {
	if isLocal(ip) {
		return &stubLocalLocation, nil
	}

	return nil, errors.New("stub geolocation service, no available location for ip")
}
