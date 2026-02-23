package aggregator

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
)

func TestCheckRequest(t *testing.T) {
	var tests = []struct {
		name    string
		method  string
		headers map[string]string
		err     error
	}{
		{"successfully accept request without headers", http.MethodGet, nil, nil},
		{"successfully accept request with Accept header", http.MethodGet, map[string]string{"Accept": "application/json"}, nil},
		{"successfully accept request with Accept-Language header", http.MethodGet, map[string]string{"Accept-Language": "en-GB"}, nil},
		{"successfully accept request with Accept-Language header array", http.MethodGet, map[string]string{"Accept-Language": "fr;en-gb"}, nil},
		{"reject request with Accept-Language header array", http.MethodGet, map[string]string{"Accept-Language": "de;en-US"}, errors.New("unsupported regional language: de;en-US")},
		{"reject request with Accept-Language header array", http.MethodGet, map[string]string{"Accept-Language": "fr;ned"}, errors.New("unsupported language: fr;ned")},
		{"reject request with Accept header", http.MethodGet, map[string]string{"Accept": "application/xml"}, errors.New("unsupported format: application/xml")},
		{"reject request with OPTIONS method header", http.MethodOptions, nil, errors.New("unsupported method: OPTIONS")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := http.Request{Method: tt.method}
			req.Header = make(http.Header)
			for key, val := range tt.headers {
				req.Header.Set(key, val)
			}
			err := checkRequest(&req)
			if err == nil && tt.err != nil {
				t.Errorf("expected error %v but received none", tt.err)
			} else if tt.err == nil && err != nil {
				t.Errorf("expected no error but received %v", err)
			} else if err != nil && tt.err != nil && err.Error() != tt.err.Error() {
				t.Errorf("expected %v, got %v", tt.err, err)
			}
		})
	}
}

func TestGetLatAndLong(t *testing.T) {
	var tests = []struct {
		name    string
		queries map[string]string
		lat     float32
		long    float32
		err     error
	}{
		{"successfully get lat and long", map[string]string{"lat": "12.34", "long": "56.78"}, 12.34, 56.78, nil},
		{"fail to get both lat and long", map[string]string{"lat": "12.34"}, 0, 0, errors.New("both lat and long are required as query parameters")},
		{"fail to get both lat and long", map[string]string{"lat": "cats", "long": "dogs"}, 0, 0, errors.New(`unable to parse lat query parameter as float32: strconv.ParseFloat: parsing "cats": invalid syntax`)},
		{"fail to parse", map[string]string{}, 0, 0, errors.New("both lat and long are required as query parameters")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := http.Request{URL: &url.URL{Path: "/"}}
			q := req.URL.Query()
			for key, val := range tt.queries {
				q.Add(key, val)
			}
			req.URL.RawQuery = q.Encode()
			lat, long, err := getLatAndLong(&req)

			if lat != tt.lat {
				t.Errorf("expected lat to be %.6f, got %.6f", tt.lat, lat)
			}
			if long != tt.long {
				t.Errorf("expected long to be %.6f, got %.6f", tt.long, long)
			}

			if err == nil && tt.err != nil {
				t.Errorf("expected error %v but received none", tt.err)
			} else if tt.err == nil && err != nil {
				t.Errorf("expected no error but received %v", err)
			} else if err != nil && tt.err != nil && err.Error() != tt.err.Error() {
				t.Errorf("expected %v, got %v", tt.err, err)
			}
		})
	}
}

func TestInitServices(t *testing.T) {
	ret := initServices(0.0, 0.0)

	if len(ret) != numServices {
		t.Errorf("expected %d service, got %d", numServices, len(ret))
	}
}
