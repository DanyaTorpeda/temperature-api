package service

import (
	"context"
	weatherapi "third-party-api"

	"github.com/redis/go-redis/v9"
)

type WeatherService struct {
	red *redis.Client
}

func NewWeatherClient(red *redis.Client) *WeatherService {
	return &WeatherService{red: red}
}

func (s *WeatherService) GetTemperature(ctx context.Context, resp *weatherapi.GeocodingResponse) {

}

func (s *WeatherService) AddTemperature(ctx context.Context) {

}
