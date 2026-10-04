package repository

import (
	"acrocuit/internal/models"
	"context"
)

type Devices interface {
	GetDevices(ctx context.Context, userEmail string, spaceId int, include string) ([]models.Device, error)
	GetDevice(ctx context.Context, userEmail string, deviceId int, include string) (*models.Device, error)
	AddDevice(ctx context.Context, userEmail string, spaceId int, name string, position *models.Position) (*models.Device, error)
	SetDevice(ctx context.Context, userEmail string, deviceId int, name *string, position *models.Position) (*models.Device, error)
	SetDevicePosition(ctx context.Context, userEmail string, deviceId int, position models.Position) (*models.Device, error)
	DelDevice(ctx context.Context, userEmail string, deviceId int) error
	GetDeviceBreakers(ctx context.Context, userEmail string, deviceId int, include string) ([]models.Breaker, error)
	AddDeviceBreaker(ctx context.Context, userEmail string, deviceId, breakerId int) error
	DelDeviceBreaker(ctx context.Context, userEmail string, deviceId, breakerId int) error
	GetDeviceUpstream(ctx context.Context, userEmail string, deviceId int) ([]models.Breaker, error)
}
