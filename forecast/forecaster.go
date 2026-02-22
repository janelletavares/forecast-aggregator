// Package forecast defines common interactions with a forecast service.
package forecast

const Days = 5

// Forecaster defines the primary method for a forecast service.
type Forecaster interface {
	GetForecast(chan map[string]WeatherOutput)
}

// ToError is a helper for communicating an error from a forecaster goroutine.
func ToError(msg string) map[string]WeatherOutput {
	m := make(map[string]WeatherOutput)
	m["error"] = WeatherOutput{ErrorMessage: msg}
	return m
}

// WeatherOutput defines the desired base set of fields in a forecast.
type WeatherOutput struct {
	Date          string  `json:"date,omitempty"`
	MaxTempC      float32 `json:"maxtemp_c,omitempty"`
	MinTempC      float32 `json:"mintemp_c,omitempty"`
	MaxWindKPH    float32 `json:"maxwind_kph,omitempty"`
	TotalPrecipMM float32 `json:"totalprecip_mm,omitempty"`
	AvgVisKM      float32 `json:"avgvis_km,omitempty"`
	AvgHumidity   int     `json:"avghumidity,omitempty"`
	ChanceOfRain  int     `json:"daily_chance_of_rain,omitempty"`
	UV            float32 `json:"uv,omitempty"`
	ErrorMessage  string  `json:"error_message,omitempty"`
}
