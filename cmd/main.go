package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const weatherAPI = "https://api.open-meteo.com/v1/forecast"
const geocodingAPI = "https://geocoding-api.open-meteo.com/v1/search"

type GeocodingResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

type WeatherResponse struct {
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
		WeatherCode int     `json:"weathercode"`
	} `json:"current_weather"`
}

func getCityName(cityName string) (*GeocodingResponse, error) {
	nameUrl := fmt.Sprintf("%s?name=%s", geocodingAPI, cityName)
	resp, err := http.Get(nameUrl)
	if err != nil {
		return nil, fmt.Errorf("couldnt make get request to open-meteo geocoding")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("couldnt read response body")
	}

	var geocodingResponse GeocodingResponse
	if err := json.Unmarshal(body, &geocodingResponse); err != nil {
		return nil, fmt.Errorf("couldnt bind response into json")
	}
	logrus.Print(geocodingResponse)

	return &geocodingResponse, nil
}

func getWeather(c *gin.Context) {
	cityName := c.Query("name")
	if cityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	geocodingResponse, err := getCityName(cityName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(geocodingResponse.Results) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city not found"})
		return
	}

	lat := geocodingResponse.Results[0].Latitude
	lon := geocodingResponse.Results[0].Longitude

	temperatureUrl := fmt.Sprintf("%s?latitude=%f&longitude=%f&current_weather=true", weatherAPI, lat, lon)

	resp, err := http.Get(temperatureUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldnt make get request to open-meteo weather API"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldnt read response body"})
		return
	}

	var weather WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldnt bind response into json"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": weather})
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.GET("/weather", getWeather)

	router.Run(":8080")
}
