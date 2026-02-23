// Package aggregator manages getting forecasts in a concurrent way.
package aggregator

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/janelletavares/forecast-aggregator/forecast"
	oma "github.com/janelletavares/forecast-aggregator/forecast/services/open_meteo_api"
	wa "github.com/janelletavares/forecast-aggregator/forecast/services/weather_api"
)

const numServices = 2
const timeout = 5 // seconds

// AggregateForecasts validates the request before fetching forecasts concurrently.
func AggregateForecasts(w http.ResponseWriter, req *http.Request) {
	if err := checkRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lat, long, err := getLatAndLong(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	services := initServices(lat, long)
	forecasts := fetchForecasts(services)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Language", "en-gb")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len([]byte(forecasts))))
	w.Header().Set("Date", time.Now().Format("Mon, 02 Jan 2006 03:04:05 MST"))
	_, err = w.Write([]byte(forecasts))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func checkRequest(req *http.Request) error {
	if req.Method != http.MethodGet {
		return fmt.Errorf("unsupported method: %v", req.Method)
	}

	// only JSON
	if accept := req.Header.Get("Accept"); accept != "" && accept != "*/*" && accept != "application/json" {
		return fmt.Errorf("unsupported format: %v", accept)
	}

	// only English with metric units
	acceptLang := req.Header.Get("Accept-Language")
	if acceptLang != "" {
		if strings.Contains(acceptLang, ",") || strings.Contains(acceptLang, ";") {
			// must check the multiple values format
			if !strings.Contains(acceptLang, "*") &&
				(strings.Contains(strings.ToLower(acceptLang), "en-us") && !strings.Contains(acceptLang, "en")) {
				return fmt.Errorf("unsupported language: %v", acceptLang)
			}
		}
		// single value format check
		if strings.Contains(strings.ToLower(acceptLang), "en-us") {
			// assuming en-US indicates a desire for Fahrenheit and inches units
			return fmt.Errorf("unsupported regional language: %v", acceptLang)
		} else if !strings.Contains(acceptLang, "en") {
			return fmt.Errorf("unsupported language: %v", acceptLang)
		}
	}

	return nil
}

func getLatAndLong(req *http.Request) (float32, float32, error) {
	var (
		lat, long       float64
		latStr, longStr string
		err             error
	)

	latStr = req.URL.Query().Get("lat")
	longStr = req.URL.Query().Get("long")
	if latStr == "" || longStr == "" {
		return 0, 0, errors.New("both lat and long are required as query parameters")
	}

	lat, err = strconv.ParseFloat(latStr, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to parse lat query parameter as float32: %v", err)
	}
	long, err = strconv.ParseFloat(longStr, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to parse long query parameter as float32: %v", err)
	}
	return float32(lat), float32(long), nil
}

func fetchForecasts(services []forecast.Forecaster) string {
	forecasts := make(chan map[string]forecast.WeatherOutput, numServices)

	for i := 0; i < numServices; i++ {
		log.Println("requesting weather forecast")
		go services[i].GetForecast(forecasts)
	}

	// gather responses
	responses := make(map[string]interface{})
	for i := 0; i < numServices; i++ {
		select {
		case msg := <-forecasts:
			log.Println("received weather forecast")
			responses[fmt.Sprintf("weatherAPI%d", i+1)] = msg
		case <-time.After(timeout * time.Second):
			log.Println("timeout waiting for weather forecast")
		}
	}

	// put forecasts into JSON response
	j, err := json.Marshal(responses)
	if err != nil {
		return fmt.Sprintf("error aggregating forecasts: %v", err)
	}

	return string(j)
}

func initServices(lat float32, long float32) []forecast.Forecaster {
	forecasters := make([]forecast.Forecaster, 0)

	forecasters = append(forecasters, wa.New(lat, long))
	forecasters = append(forecasters, oma.New(lat, long))

	return forecasters
}
