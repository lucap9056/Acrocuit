package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Devices interface {
	GetDevices(ctx context.Context, userEmail string, spaceId models.SpaceId, include string) ([]models.Device, error)
	GetDevice(ctx context.Context, userEmail string, deviceId models.DeviceId, include string) (*models.Device, error)
	AddDevice(ctx context.Context, userEmail string, spaceId models.SpaceId, name string, position *models.Position, attributes models.Attributes) (*models.Device, error)
	SetDevice(ctx context.Context, userEmail string, deviceId models.DeviceId, name *string, position *models.Position, attributes models.Attributes) (*models.Device, error)
	SetDevicePosition(ctx context.Context, userEmail string, deviceId models.DeviceId, position models.Position) (*models.Device, error)
	DelDevice(ctx context.Context, userEmail string, deviceId models.DeviceId) error
	GetDeviceSwitches(ctx context.Context, userEmail string, deviceId models.DeviceId, include string) ([]models.LinkedSwitch, error)
	AddDeviceSwitch(ctx context.Context, userEmail string, deviceId models.DeviceId, switchId models.SwitchId, attributes models.Attributes) (*models.DeviceLink, error)
	SetDeviceSwitch(ctx context.Context, userEmail string, deviceId models.DeviceId, switchId models.SwitchId, attributes models.Attributes) (*models.DeviceLink, error)
	DelDeviceSwitch(ctx context.Context, userEmail string, deviceId models.DeviceId, switchId models.SwitchId) error
	GetDeviceUpstream(ctx context.Context, userEmail string, deviceId models.DeviceId) (*models.PowerChainGraph, error)
}
