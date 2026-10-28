package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Devices interface {
	GetDevices(ctx context.Context, userEmail string, spaceId string, include string) ([]models.Device, error)
	GetDevice(ctx context.Context, userEmail string, deviceId string, include string) (*models.Device, error)
	AddDevice(ctx context.Context, userEmail string, spaceId string, name string, position *models.Position, attributes models.Attributes) (*models.Device, error)
	SetDevice(ctx context.Context, userEmail string, deviceId string, name *string, position *models.Position, attributes models.Attributes) (*models.Device, error)
	SetDevicePosition(ctx context.Context, userEmail string, deviceId string, position models.Position) (*models.Device, error)
	DelDevice(ctx context.Context, userEmail string, deviceId string) error
	GetDeviceSwitches(ctx context.Context, userEmail string, deviceId string, include string) ([]models.LinkedSwitch, error)
	AddDeviceSwitch(ctx context.Context, userEmail string, deviceId, switchId string, attributes models.Attributes) (*models.DeviceLink, error)
	SetDeviceSwitch(ctx context.Context, userEmail string, deviceId, switchId string, attributes models.Attributes) (*models.DeviceLink, error)
	DelDeviceSwitch(ctx context.Context, userEmail string, deviceId, switchId string) error
	GetDeviceUpstream(ctx context.Context, userEmail string, deviceId string) (*models.PowerChain, error)
}
