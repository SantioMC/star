package geolocation

import (
	"context"
	"fmt"
)

type Service interface {
	GetLocation(ctx context.Context, ip string) (*Geolocation, error)
}

type Geolocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Region    string  `json:"region"`
	ISP       string  `json:"isp"`
}

func (g *Geolocation) String() string {
	return fmt.Sprintf("%s (%s, %s)", g.Country, g.City, g.Region)
}

func GetService() Service {
	return &stubGeolocationService{}
}
