package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/tubblelive/star/internal/geolocation"
)

type GeoserviceConfig struct {
	Type string `env:"GEOSERVICE,disabled"`
	Oval OvalConfig
}

type OvalConfig struct {
	URL *string `env:"OVAL_URL"`
}

func (c *GeoserviceConfig) IsEnabled() bool {
	return c.Type != "disabled"
}

func (c *GeoserviceConfig) CreateService() geolocation.Service {
	switch strings.ToLower(c.Type) {
	case "oval":
		if c.Oval.URL == nil {
			slog.Error("oval geolocation service was specified, required OVAL_URL environment variable not set")
			os.Exit(1)
		}

		return &geolocation.OvalService{
			Url: *c.Oval.URL,
		}
	default:
		return &geolocation.StubGeolocationService{}
	}
}
