package geolocation

import "context"

type stubGeolocationService struct{}

func (s *stubGeolocationService) GetLocation(ctx context.Context, ip string) (*Geolocation, error) {
	return &Geolocation{
		Latitude:  0.000,
		Longitude: 0.000,
		Country:   "US",
		City:      "",
		Region:    "",
		ISP:       "",
	}, nil
}
