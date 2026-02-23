//go:build integration

package aggregator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/janelletavares/forecast-aggregator/forecast"
)

// TestAggregateForecasts makes HTTP requests.
func TestAggregateForecasts(t *testing.T) {
	w := httptest.NewRecorder()
	req := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "/weather"}}
	q := req.URL.Query()
	q.Add("lat", "12.34")
	q.Add("long", "45.67")
	req.URL.RawQuery = q.Encode()
	err := os.Setenv("WEATHER_API_KEY", "CHANGEME")
	if err != nil {
		t.Fatalf("unexpected error setting up test: %v", err)
	}
	AggregateForecasts(w, req)

	bytes, err := io.ReadAll(w.Body)
	if err != nil {
		fmt.Errorf("unexpected error when getting response body: %v", err)
	}
	fmt.Printf("output: %+v\n", w.Result())
	fmt.Printf("body: %v\n", string(bytes))
}

func TestFetchForecasts(t *testing.T) {
	parsed := make(map[string]interface{})

	output := fetchForecasts(getFakeServices())
	err := json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("unexpected output format: %v %v", err, output)
	}
	fmt.Println(output)

	for _, key := range []string{"weatherAPI1", "weatherAPI2"} {
		_, ok := parsed[key]
		if !ok {
			t.Errorf("expected key %q, but not present", key)
		}
	}
}

type serv struct {
}

func (s *serv) GetForecast(c chan map[string]forecast.WeatherOutput) {
	output := make(map[string]forecast.WeatherOutput)
	output["it's raining men today"] = forecast.WeatherOutput{
		Date: "today",
	}
	output["it's raining men tomorrow"] = forecast.WeatherOutput{
		Date: "tomorrow",
	}
	c <- output
}

func getFakeServices() []forecast.Forecaster {
	one := serv{}
	two := serv{}
	return []forecast.Forecaster{&one, &two}
}

func TestFetchForecasts_Timeout(t *testing.T) {
	defer func() { _ = recover() }()
	parsed := make(map[string]interface{})

	output := fetchForecasts(getSlowServices())
	err := json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("unexpected output format: %v %v", err, output)
	}
	fmt.Println(output)

	if len(parsed) != 1 {
		t.Errorf("expecting exactly 1 forecast, found %d", len(output))
	}
	key := "weatherAPI1"
	_, ok := parsed[key]
	if !ok {
		t.Errorf("expected key %q, but not present", key)
	}
}

type slowServ struct {
}

func (s *slowServ) GetForecast(c chan map[string]forecast.WeatherOutput) {
	output := make(map[string]forecast.WeatherOutput)
	time.Sleep(6 * time.Second)
	output["it's raining men today"] = forecast.WeatherOutput{
		Date: "today",
	}
	output["it's raining men tomorrow"] = forecast.WeatherOutput{
		Date: "tomorrow",
	}
	c <- output
}

func getSlowServices() []forecast.Forecaster {
	one := slowServ{}
	two := serv{}
	return []forecast.Forecaster{&one, &two}
}
