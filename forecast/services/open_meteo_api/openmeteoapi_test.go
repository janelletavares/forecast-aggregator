package open_meteo_api

import (
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/janelletavares/forecast-aggregator/forecast"
)

const sampleOutput = `{"latitude":52.52,"longitude":13.419998,"generationtime_ms":37.937283515930176,"utc_offset_seconds":0,"timezone":"GMT","timezone_abbreviation":"GMT","elevation":38.0,"daily_units":{"time":"iso8601","temperature_2m_max":"°C","temperature_2m_min":"°C","apparent_temperature_max":"°C","apparent_temperature_min":"°C","uv_index_max":"","precipitation_sum":"mm","wind_speed_10m_max":"km/h","wind_direction_10m_dominant":"°","cloud_cover_mean":"%","relative_humidity_2m_mean":"%","precipitation_probability_mean":"%","visibility_mean":"m","wind_speed_10m_min":"km/h"},"daily":{"time":["2026-02-18","2026-02-19","2026-02-20","2026-02-21","2026-02-22"],"temperature_2m_max":[-0.6,0.1,0.6,8.2,10.8],"temperature_2m_min":[-3.7,-6.8,-6.6,-0.4,6.6],"apparent_temperature_max":[-4.2,-4.7,-4.6,5.1,8.8],"apparent_temperature_min":[-7.4,-11.3,-11.1,-4.6,4.0],"uv_index_max":[2.30,2.20,2.35,0.80,0.25],"precipitation_sum":[0.00,0.00,0.00,3.10,7.30],"wind_speed_10m_max":[9.0,12.5,15.8,14.8,17.1],"wind_direction_10m_dominant":[12,83,129,229,246],"cloud_cover_mean":[77,68,78,94,99],"relative_humidity_2m_mean":[76,67,68,85,90],"precipitation_probability_mean":[3,0,0,31,66],"visibility_mean":[26684.17,36800.00,34527.50,25190.83,16759.17],"wind_speed_10m_min":[2.7,5.9,6.7,9.4,6.9]}}`

func TestGetFiveDayForecast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleOutput))
	}))
	defer server.Close()
	s := &service{lat: 12.34, long: 56.78, getURL: func(a float32, b float32) string { return server.URL }}
	c := make(chan map[string]forecast.WeatherOutput, 1)

	s.GetForecast(c)
	value := <-c

	if len(value) != 5 {
		t.Errorf("Expected %d forecasts, got %d", forecast.Days, len(value))
	}
}

func TestGetFiveDayForecast_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(``))
	}))
	defer server.Close()
	s := &service{lat: 12.34, long: 56.78, getURL: func(a float32, b float32) string { return server.URL }}
	c := make(chan map[string]forecast.WeatherOutput, 1)

	s.GetForecast(c)
	value := <-c

	if len(value) != 1 {
		t.Errorf("Expected response with error message, got %+v", value)
	}
	msg := value["error"].ErrorMessage
	if !strings.Contains(msg, "error processing http response") {
		t.Errorf("Expected error message, got %+v", msg)
	}
}

func TestGetURL(t *testing.T) {
	var tests = []struct {
		name     string
		lat      float32
		long     float32
		contains []string
	}{
		{"successfully accept negative coordinates", 10.01, 34.560001, []string{"latitude=10.010000", "longitude=34.560001"}},
		{"successfully accept positive coordinates", -10.01, -34.560001, []string{"latitude=-10.010000", "longitude=-34.560001"}},
	}

	currentTime := time.Now().Local()
	today := currentTime.Format("2006-01-02")
	fifthDay := currentTime.Add(4 * 24 * time.Hour).Format("2006-01-02")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := getURL(tt.lat, tt.long)
			if !strings.Contains(output, baseURL) {
				t.Errorf("expected %s in the URL: %s", baseURL, output)
			}
			if !strings.Contains(output, fmt.Sprintf("start_date=%s", today)) {
				t.Errorf("expected start_date to be %s in URL: %s", today, output)
			}
			if !strings.Contains(output, fmt.Sprintf("end_date=%s", fifthDay)) {
				t.Errorf("expected end_date to be %s in URL: %s", fifthDay, output)
			}
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("expected %s in the URL: %s", s, output)
				}
			}
		})
	}
}

func TestConvert(t *testing.T) {
	var tests = []struct {
		name  string
		input apiResponse
		want  map[string]forecast.WeatherOutput
	}{
		{"Empty should convert to empty", apiResponse{}, map[string]forecast.WeatherOutput{}},
		{name: "Populated response should convert to populated forecast", input: apiResponse{
			Daily: daily{
				Time:                  []string{"one", "two"},
				MaxTemp:               []float32{1.2, 3.4},
				MinTemp:               []float32{1.2, 3.4},
				UV:                    []float32{2.2, 3.3},
				MaxWindSpeed:          []float32{1.1, 5.5},
				Precipitation:         []float32{2.2, 3.3},
				RelativeHumidity:      []int{40, 50},
				ChanceOfPrecipitation: []int{10, 20},
				AvgVis:                []float32{10.10, 20.20},
			}}, want: map[string]forecast.WeatherOutput{
			"day1": forecast.WeatherOutput{
				Date:          "one",
				MaxTempC:      1.2,
				MinTempC:      1.2,
				MaxWindKPH:    1.1,
				TotalPrecipMM: 2.2,
				AvgVisKM:      10.10,
				AvgHumidity:   40,
				ChanceOfRain:  10,
				UV:            2.2,
			},
			"day2": forecast.WeatherOutput{
				Date:          "two",
				MaxTempC:      3.4,
				MinTempC:      3.4,
				MaxWindKPH:    5.5,
				TotalPrecipMM: 3.3,
				AvgVisKM:      20.20,
				AvgHumidity:   50,
				ChanceOfRain:  20,
				UV:            3.3,
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := convert(&tt.input)
			if !maps.Equal(output, tt.want) {
				t.Errorf("Expected %v, got %v", tt.want, output)
			}
		})
	}
}
