package model

import "github.com/zitadel/oidc/v3/pkg/op"

type DeviceAuthorizationEntry struct {
	DeviceCode string
	UserCode   string
	State      *op.DeviceAuthorizationState
}