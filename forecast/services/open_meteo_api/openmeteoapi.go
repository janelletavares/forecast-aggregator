// Package open_meteo_api makes forecast requests.
package open_meteo_api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/janelletavares/forecast-aggregator/forecast"
)

// New creates a service that makes forecast requests from Open Meteo API.
func New(lat float32, long float32) forecast.Forecaster {
	return &service{lat: lat, long: long, getURL: getURL}
}

// GetForecast fetches a future forecast from Open Meteo API.
func (s *service) GetForecast(c chan map[string]forecast.WeatherOutput) {
	log.Println("making Open Meteo API call")
	res, err := http.Get(s.getURL(s.lat, s.long))
	log.Println("received Open Meteo API response")
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

type daily struct {
	Time                  []string  `json:"time"`
	MaxTemp               []float32 `json:"temperature_2m_max"`
	MinTemp               []float32 `json:"temperature_2m_min"`
	UV                    []float32 `json:"uv_index_max"`
	MaxWindSpeed          []float32 `json:"wind_speed_10m_max"`
	Precipitation         []float32 `json:"precipitation_sum"`
	RelativeHumidity      []int     `json:"relative_humidity_2m_mean"`
	ChanceOfPrecipitation []int     `json:"precipitation_probability_mean"`
	AvgVis                []float32 `json:"visibility_mean"`
}

type apiResponse struct {
	Daily daily `json:"daily"`
}

const baseURL = "https://api.open-meteo.com/v1/"

func getURL(lat float32, long float32) string {
	currentTime := time.Now().Local()
	today := currentTime.Format("2006-01-02")
	fifthDay := currentTime.Add((forecast.Days - 1) * 24 * time.Hour).Format("2006-01-02")
	return fmt.Sprintf("%sforecast?latitude=%f&longitude=%f&start_date=%s&end_date=%s&daily=temperature_2m_max,temperature_2m_min,uv_index_max,precipitation_sum,wind_speed_10m_max,relative_humidity_2m_mean,precipitation_probability_mean,visibility_mean", baseURL, lat, long, today, fifthDay)
}

// convert takes the Open Meteo API response format and transforms to this project's format.
func convert(resp *apiResponse) map[string]forecast.WeatherOutput {
	ret := make(map[string]forecast.WeatherOutput)
	l := len(resp.Daily.Time)
	for i := 0; i < l; i++ {
		key := fmt.Sprintf("day%d", i+1)
		val := forecast.WeatherOutput{
			Date:          resp.Daily.Time[i],
			MaxTempC:      resp.Daily.MaxTemp[i],
			MinTempC:      resp.Daily.MinTemp[i],
			MaxWindKPH:    resp.Daily.MaxWindSpeed[i],
			TotalPrecipMM: resp.Daily.Precipitation[i],
			AvgVisKM:      resp.Daily.AvgVis[i],
			AvgHumidity:   resp.Daily.RelativeHumidity[i],
			ChanceOfRain:  resp.Daily.ChanceOfPrecipitation[i],
			UV:            resp.Daily.UV[i],
		}
		ret[key] = val
	}
	return ret
}
