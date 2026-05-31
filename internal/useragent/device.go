package useragent

import (
	"github.com/medama-io/go-useragent/agents"
)

type DeviceType string

const (
	DeviceDesktop DeviceType = "desktop"
	DeviceBot     DeviceType = "bot"
	DeviceTablet  DeviceType = "tablet"
	DeviceTV      DeviceType = "tv"
	DeviceMobile  DeviceType = "mobile"
	DeviceUnknown DeviceType = "unknown"
)

func GetDevice(ua string) DeviceType {
	agent := parser.Parse(ua)

	switch agent.Device() {
	case agents.DeviceDesktop:
		return DeviceDesktop
	case agents.DeviceTablet:
		return DeviceTablet
	case agents.DeviceTV:
		return DeviceTV
	case agents.DeviceMobile:
		return DeviceMobile
	case agents.DeviceBot:
		return DeviceBot
	default:
		return DeviceUnknown
	}
}
