package geolocation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type OvalService struct {
	Url string
}

type OvalGeolocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Continent string  `json:"continent"`
	Province  string  `json:"province"`
}

type OvalAddressLookup struct {
	Address string `json:"address"`
}

func (service *OvalService) GetLocation(ctx context.Context, ip string) (*string, error) {
	if isLocal(ip) {
		slog.Warn("local ip was provided to oval geoservice", "ip", ip)
		return &stubLocalLocation, nil
	}

	// Simple one-ip cidr builder, we should only ever get passed real ips here so we do a very simple check to determine between
	// ipv4 and ipv6
	cidr := 32
	if !strings.Contains(ip, ".") {
		cidr = 128
	}

	data, err := json.Marshal(OvalAddressLookup{
		Address: fmt.Sprintf("%s/%d", ip, cidr),
	})

	if err != nil {
		return nil, err
	}

	res, err := http.Post(
		fmt.Sprintf("%s/query", service.Url),
		"application/json",
		bytes.NewReader(data),
	)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var geo OvalGeolocation
	if err := json.Unmarshal(bytes, &geo); err != nil {
		return nil, err
	}

	value := fmt.Sprintf("%s (%s, %s)", geo.Country, geo.City, geo.Province)
	return &value, nil
}
