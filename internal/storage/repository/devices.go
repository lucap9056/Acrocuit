package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Devices interface {
	GetDevices(ctx context.Context, userEmail string, spaceId int, include string) ([]models.Device, error)
	GetDevice(ctx context.Context, userEmail string, deviceId int, include string) (*models.Device, error)
	AddDevice(ctx context.Context, userEmail string, spaceId int, name string, position *models.Position, attributes models.Attributes) (*models.Device, error)
	SetDevice(ctx context.Context, userEmail string, deviceId int, name *string, position *models.Position, attributes models.Attributes) (*models.Device, error)
	SetDevicePosition(ctx context.Context, userEmail string, deviceId int, position models.Position) (*models.Device, error)
	DelDevice(ctx context.Context, userEmail string, deviceId int) error
	GetDeviceSwitches(ctx context.Context, userEmail string, deviceId int, include string) ([]models.LinkedSwitch, error)
	AddDeviceSwitch(ctx context.Context, userEmail string, deviceId, switchId int, attributes models.Attributes) (*models.DeviceLink, error)
	SetDeviceSwitch(ctx context.Context, userEmail string, deviceId, switchId int, attributes models.Attributes) (*models.DeviceLink, error)
	DelDeviceSwitch(ctx context.Context, userEmail string, deviceId, switchId int) error
	GetDeviceUpstream(ctx context.Context, userEmail string, deviceId int) (*models.PowerChain, error)
}
