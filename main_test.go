package main

import (
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/recorder"
)

func TestGetWeatherReport(t *testing.T) {
	end := vcr("fixtures/get-weather-report")
	defer end()

	weather, err := getWeatherReport("Copenhagen", "DK")
	if err != nil {
		log.Fatal(err)
	}

	if weather.Title != "Copenhagen, Capital (Denmark)" {
		log.Fatal("Unexpected weather page title: ", weather.Title)
	}

	for _, d := range weather.Days {
		for _, p := range d.Periods {
			if !strings.Contains(p.Time, ":") {
				log.Fatal("Time should contain ':'")
			}

			if !strings.Contains(p.Temp, "°") {
				log.Fatal("Temp should contain '°'")
			}

			if !strings.Contains(p.Precipitation, "mm") {
				log.Fatal("Precipitation should contain 'mm'")
			}
		}
	}
}

func TestGetWeatherReportKBHSV(t *testing.T) {
	end := vcr("fixtures/get-weather-report-kbh-sv")
	defer end()

	weather, err := getWeatherReport("Kobenhavn SV", "DK")
	if err != nil {
		log.Fatal(err)
	}

	if weather.Title != "Copenhagen, Capital (Denmark)" {
		log.Fatal("Unexpected weather page title: ", weather.Title)
	}
}

func TestGetLocation(t *testing.T) {
	end := vcr("fixtures/get-location")
	defer end()

	l, err := getLocation()
	if err != nil {
		log.Fatal(err)
	}

	if l.CountryCode != "DK" || l.City != "Kobenhavn SV" {
		log.Fatal("Unexpected location", l.CountryCode, l.City)
	}
}

func vcr(name string) func() {
	recorder, err := recorder.New(name)
	if err != nil {
		panic(err)
	}

	oldTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = recorder
	return func() {
		http.DefaultClient.Transport = oldTransport
		recorder.Stop()
	}
}
