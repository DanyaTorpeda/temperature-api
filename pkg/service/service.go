package service

import (
	"context"
	weatherapi "third-party-api"

	"github.com/redis/go-redis/v9"
)

type CityCoordinates interface {
	GetCityCoordinates(context.Context, string) (*weatherapi.GeocodingResponse, error)
	AddCityCoordinates(context.Context, string, weatherapi.GeocodingResponse) (*weatherapi.GeocodingResponse, error)
}

type Weather interface {
	GetTemperature(context.Context, *weatherapi.GeocodingResponse)
}

type Service struct {
	CityCoordinates
	Weather
}

func NewService(client *redis.Client) *Service {
	return &Service{
		CityCoordinates: NewCoordinatesService(client),
		Weather:         NewWeatherClient(client),
	}
}
