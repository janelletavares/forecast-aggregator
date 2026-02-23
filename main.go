// Package main handles the HTTP server, routing, and logging.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/janelletavares/forecast-aggregator/forecast/aggregator"
)

const timeFormat = time.StampNano

type logWriter struct{}

func (lw *logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format(timeFormat), "Z | ", string(bytes))
}

func main() {
	log.SetFlags(0) // Omit default prefixes.
	log.SetOutput(new(logWriter))

	http.HandleFunc("/weather", aggregator.AggregateForecasts)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("unexpected error from ListenAndServe: %v", err)
	}
}
