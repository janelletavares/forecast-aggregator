# Overview of the Forecast Aggregator

The repo describes a service that aggregates two forecast APIs - Open Meteo and Weather API.

This is accomplished via an HTTP endpoint request at localhost:8080/weather. There is only one endpoint, and it only accepts the GET method; HEAD and OPTIONS methods are not supported. The GET request will fetch a 5-day forecast from two APIs in parallel and combine them into one response. This service only processes forecast data in JSON and in English.


## Prerequisites

- Install Go 1.25
- Set WEATHER_API_KEY environment variable from the [provided secret](https://send.bitwarden.com/#2S3BU51Qv0GnkLP4AUG07g/IUj7PUJ2NlY-b6PwP6V_RQ)
- Install cURL or another means to make HTTP requests


## Running

To do a sanity check that most pre-requisites are in place, run `go build ./...`

In one shell, start the server with `go run main.go`

In another shell or program, make HTTP requests, such as:
```
curl -X GET -v \
-H "Accept-Language: en-gb" \
-H "Accept: application/json" \
"localhost:8080/weather?lat=52.07667&long=4.29861"
```

And for an easy-to-read version, one could try:
```
curl -X GET "localhost:8080/weather?lat=52.07667&long=4.29861" > response.json
cat response.json | python3 -m json.tool
```


## Testing

In order to run unit tests, issue this command: 
`go test -v ./...` 

In order to run the unit tests and the integration tests, which works best with updating the appropriate env var, issue this command:
`go test -tags=integration ./...`

## Extensibility

In order to extend this service to include additional weather APIs, start by defining the additional API wrapper under the package forecast/services. Be sure to follow the same interface as the other existing wrappers. Then, increment numServices in the aggregator package, and append the new service to the slice of forecasters in the `initServices` function in the same package. 
