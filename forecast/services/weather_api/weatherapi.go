// Package weather_api makes forecast requests.
package weather_api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/janelletavares/forecast-aggregator/forecast"
)

// New creates a service that makes forecast requests from Open Meteo API.
func New(lat float32, long float32) forecast.Forecaster {
	return &service{lat: lat, long: long, getURL: getURL}
}

// GetForecast fetches a future forecast from Weather API.
func (s *service) GetForecast(c chan map[string]forecast.WeatherOutput) {
	log.Println("making Weather API request")
	res, err := http.Get(s.getURL(s.lat, s.long))
	log.Println("received Weather API response")
	if err != nil {
		c <- forecast.ToError(fmt.Sprintf("error making http request: %s\n", err))
		return
	}
	var r apiResponse
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		c <- forecast.ToError(fmt.Sprintf("error processing http response: %s\n", err))
		return
	}
	c <- convert(&r)
}

type service struct {
	lat    float32
	long   float32
	getURL func(float32, float32) string
}

type day struct {
	MaxTempC      float32 `json:"maxtemp_c"`
	MinTempC      float32 `json:"mintemp_c"`
	MaxWindKPH    float32 `json:"maxwind_kph"`
	TotalPrecipMM float32 `json:"totalprecip_mm"`
	AvgVisKM      float32 `json:"avgvis_km"`
	AvgHumidity   int     `json:"avghumidity"`
	ChanceOfRain  int     `json:"daily_chance_of_rain"`
	UV            float32 `json:"uv"`
}

type forecastDay struct {
	Date string `json:"date"`
	Day  day    `json:"day"`
}

type days struct {
	ForecastDay []forecastDay `json:"forecastday"`
}

type apiResponse struct {
	Forecast days `json:"forecast"`
}

const baseURL = "https://api.weatherapi.com/v1/"
const apiKeyEnvVar = "WEATHER_API_KEY"

func getURL(lat float32, long float32) string {
	key := os.Getenv(apiKeyEnvVar)
	if key == "" {
		log.Panic("WEATHER_API_KEY environment variable is required")
	}
	return fmt.Sprintf("%sforecast.json?key=%s&q=%f,%f&days=%d", baseURL, key, lat, long, forecast.Days)
}

// convert takes the Weather API response format and transforms to this project's format.
func convert(resp *apiResponse) map[string]forecast.WeatherOutput {
	ret := make(map[string]forecast.WeatherOutput)
	for i, day := range resp.Forecast.ForecastDay {
		key := fmt.Sprintf("day%d", i+1)
		val := forecast.WeatherOutput{
			Date:          day.Date,
			MaxTempC:      day.Day.MaxTempC,
			MinTempC:      day.Day.MinTempC,
			MaxWindKPH:    day.Day.MaxWindKPH,
			TotalPrecipMM: day.Day.TotalPrecipMM,
			AvgVisKM:      day.Day.AvgVisKM,
			AvgHumidity:   day.Day.AvgHumidity,
			ChanceOfRain:  day.Day.ChanceOfRain,
			UV:            day.Day.UV,
		}
		ret[key] = val
	}
	return ret
}
