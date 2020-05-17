package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/recorder"
)

var r *recorder.Recorder

func init() {
	var err error
	r, err = recorder.New("fixtures/yr")
	if err != nil {
		log.Fatal(err)
	}

	http.DefaultClient.Transport = r
}

func shutdown() {
	r.Stop()
}

func TestMain(m *testing.M) {
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestGetWeatherReport(t *testing.T) {
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
	weather, err := getWeatherReport("Kobenhavn SV", "DK")
	if err != nil {
		log.Fatal(err)
	}

	if weather.Title != "Copenhagen, Capital (Denmark)" {
		log.Fatal("Unexpected weather page title: ", weather.Title)
	}
}

func TestGetLocation(t *testing.T) {
	l, err := getLocation()
	if err != nil {
		log.Fatal(err)
	}

	if l.CountryCode != "DK" || l.City != "Kobenhavn SV" {
		log.Fatal("Unexpected location", l.CountryCode, l.City)
	}
}
