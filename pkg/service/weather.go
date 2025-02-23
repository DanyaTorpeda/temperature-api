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
	tempPrefix         = "temperature"
	temperature        = "Temperature"
	windSpeed          = "Windspeed"
	weatherCode        = "Weathercode"
	tempExpirationTime = 1 * time.Hour
)

type WeatherService struct {
	red *redis.Client
}

func NewWeatherClient(red *redis.Client) *WeatherService {
	return &WeatherService{red: red}
}

func (s *WeatherService) GetTemperature(ctx context.Context, cityName string) (*weatherapi.WeatherResponse, error) {
	logrus.Print("inside GetTemperature")
	temp, err := s.red.HGet(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, tempPrefix), temperature).Float64()
	if errors.Is(err, redis.Nil) {
		return nil, redis.Nil
	}
	if err != nil {
		return nil, err
	}

	wSpeed, err := s.red.HGet(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, tempPrefix), windSpeed).Float64()
	if errors.Is(err, redis.Nil) {
		return nil, redis.Nil
	}
	if err != nil {
		return nil, err
	}

	wCode, err := s.red.HGet(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, tempPrefix), weatherCode).Int()
	if errors.Is(err, redis.Nil) {
		return nil, redis.Nil
	}
	if err != nil {
		return nil, err
	}

	weatherResponse := &weatherapi.WeatherResponse{CurrentWeather: struct {
		Temperature float64 "json:\"temperature\""
		Windspeed   float64 "json:\"windspeed\""
		WeatherCode int     "json:\"weathercode\""
	}{
		temp,
		wSpeed,
		wCode,
	}}

	return weatherResponse, nil
}

func (s *WeatherService) AddTemperature(ctx context.Context, cityName string, resp *weatherapi.WeatherResponse) (*weatherapi.WeatherResponse, error) {
	logrus.Print("inside AddTemperature")
	_, err := s.red.HSet(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, tempPrefix),
		"Temperature", resp.CurrentWeather.Temperature,
		"Windspeed", resp.CurrentWeather.Windspeed,
		"Weathercode", resp.CurrentWeather.WeatherCode).Result()
	if err != nil {
		return nil, err
	}
	err = s.red.Expire(ctx, fmt.Sprintf("%s:%s:%s", cityPrefix, cityName, tempPrefix), tempExpirationTime).Err()
	if err != nil {
		return nil, err
	}

	weatherResponse := &weatherapi.WeatherResponse{CurrentWeather: struct {
		Temperature float64 "json:\"temperature\""
		Windspeed   float64 "json:\"windspeed\""
		WeatherCode int     "json:\"weathercode\""
	}{
		resp.CurrentWeather.Temperature,
		resp.CurrentWeather.Windspeed,
		resp.CurrentWeather.WeatherCode,
	}}
	return weatherResponse, nil
}
