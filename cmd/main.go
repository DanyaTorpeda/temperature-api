package main

import (
	weatherapi "third-party-api"
	"third-party-api/pkg/handler"
	"third-party-api/pkg/service"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	services := service.NewService(redisClient)
	handlers := handler.NewHandler(services)
	srv := new(weatherapi.Server)
	if err := srv.Run("8080", handlers.InitRoutes()); err != nil {
		logrus.Errorf("error occured running server: %s", err.Error())
	}
}
