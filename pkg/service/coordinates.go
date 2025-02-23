package service

import (
	"context"
	"errors"
	"fmt"
	weatherapi "third-party-api"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	cityPrefix          = "city"
	coordSufix          = "coord"
	latitude            = "Latitude"
	longitude           = "Longitude"
	coordExpirationTime = time.Hour * 24 * 30
)

type CoordinatesService struct {
	red *redis.Client
}

func NewCoordinatesService(client *redis.Client) *CoordinatesService {
	return &CoordinatesService{red: client}
}

func (s *CoordinatesService) GetCityCoordinates(ctx context.Context, cityName string) (*weatherapi.GeocodingResponse, error) {
	logrus.Print("inside GetCityCoordinates")
	lat, err := s.red.HGet(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, coordSufix), latitude).Float64()
	if errors.Is(err, redis.Nil) {
		return nil, redis.Nil
	}
	if err != nil {
		return nil, err
	}
	lon, err := s.red.HGet(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, coordSufix), longitude).Float64()
	if errors.Is(err, redis.Nil) {
		return nil, redis.Nil
	}
	if err != nil {
		return nil, err
	}

	geocodingResponse := &weatherapi.GeocodingResponse{Results: make([]struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}, 1)}
	geocodingResponse.Results[0].Latitude = lat
	geocodingResponse.Results[0].Longitude = lon

	return geocodingResponse, nil
}

func (s *CoordinatesService) AddCityCoordinates(ctx context.Context, cityName string, geoResp weatherapi.GeocodingResponse) (*weatherapi.GeocodingResponse, error) {
	logrus.Print("inside AddCityCoordinates")
	_, err := s.red.HSet(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, coordSufix),
		"Latitude", geoResp.Results[0].Latitude,
		"Longitude", geoResp.Results[0].Longitude,
	).Result()
	if err != nil {
		return nil, err
	}
	if err := s.red.Expire(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, coordSufix), coordExpirationTime).Err(); err != nil {
		return nil, err
	}

	geocodingResponse := &weatherapi.GeocodingResponse{Results: make([]struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}, 1)}
	geocodingResponse.Results[0].Latitude = geoResp.Results[0].Latitude
	geocodingResponse.Results[0].Longitude = geoResp.Results[0].Longitude

	return geocodingResponse, nil
}
