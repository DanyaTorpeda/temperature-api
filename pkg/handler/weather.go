package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	weatherapi "third-party-api"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	geocodingAPI = "https://geocoding-api.open-meteo.com/v1/search"
	weatherAPI   = "https://api.open-meteo.com/v1/forecast"
)

func (h *Handler) getWeather(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cityName := c.Query("name")
	if cityName == "" {
		newErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid city name data").Error())
		return
	}

	coords, err := h.service.CityCoordinates.GetCityCoordinates(ctx, cityName)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			url := fmt.Sprintf("%s?name=%s", geocodingAPI, cityName)
			resp, err := http.Get(url)
			if err != nil {
				newErrorResponse(c, http.StatusBadRequest, err.Error())
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
				return
			}

			var geocodingResponse weatherapi.GeocodingResponse
			if err := json.Unmarshal(body, &geocodingResponse); err != nil {
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
				return
			}

			coords, err = h.service.CityCoordinates.AddCityCoordinates(ctx, cityName, geocodingResponse)
			if err != nil {
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
				return
			}
			err = nil
		} else {
			logrus.Print("Unexpected error from Redis")
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	weather, err := h.service.Weather.GetTemperature(ctx, cityName)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			url := fmt.Sprintf("%s?latitude=%f&longitude=%f&current_weather=true",
				weatherAPI, coords.Results[0].Latitude, coords.Results[0].Longitude)

			resp, err := http.Get(url)
			if err != nil {
				newErrorResponse(c, http.StatusBadRequest, err.Error())
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
				return
			}

			var weatherResponse weatherapi.WeatherResponse
			if err := json.Unmarshal(body, &weatherResponse); err != nil {
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
				return
			}

			weather, err = h.service.Weather.AddTemperature(ctx, cityName, &weatherResponse)
			if err != nil {
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
				return
			}

			err = nil
		} else {
			logrus.Print("Unexpected error from Redis")
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	weather.Latitude = coords.Results[0].Latitude
	weather.Longitude = coords.Results[0].Longitude

	c.JSON(http.StatusOK, weather)
}

// url := fmt.Sprintf("%s?latitude=%f&longitude=%f&current_weather=true",
// 	weatherAPI, coords.Results[0].Latitude, coords.Results[0].Longitude)

// resp, err := http.Get(url)
// if err != nil {
// 	newErrorResponse(c, http.StatusBadRequest, err.Error())
// 	return
// }
// defer resp.Body.Close()

// body, err := io.ReadAll(resp.Body)
// if err != nil {
// 	newErrorResponse(c, http.StatusInternalServerError, err.Error())
// 	return
// }

// var weatherResponse weatherapi.WeatherResponse
// if err := json.Unmarshal(body, &weatherResponse); err != nil {
// 	newErrorResponse(c, http.StatusInternalServerError, err.Error())
// 	return
// }

// c.JSON(http.StatusOK, weatherResponse)
