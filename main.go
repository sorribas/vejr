package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/olekukonko/tablewriter"
)

func main() {
	loc, err := getLocation()
	if err != nil {
		fmt.Println("Couldn't determine your location. Please provide the city where you are on the command line.", err)
		return
	}

	weather, err := getWeatherReport(loc.City, loc.CountryCode)
	if err != nil {
		fmt.Println("Error getting weather report:", err)
		fmt.Println("Try updating vejr")
		return
	}

	for _, day := range weather {
		fmt.Println(day.Title)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Time", "Forecast", "Temp", "Precipitation", "Wind"})

		for _, period := range day.Periods {
			table.Append([]string{period.Time, period.Forecast, period.Temp, period.Precipitation, period.Wind})
		}

		table.Render()
		fmt.Println()
	}

	getLocation()
}

func weatherReportFromDocument(doc *goquery.Document) []WeatherDay {
	result := []WeatherDay{}
	doc.Find(".yr-table-overview2").Each(func(i int, table *goquery.Selection) {
		day := WeatherDay{}
		day.Title = strings.TrimSpace(table.Find("caption").Text())

		table.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
			period := WeatherPeriod{}
			period.Time = row.Find("td:nth-child(1)").Text()
			period.Forecast = row.Find("td:nth-child(2) figcaption").Text()
			period.Temp = row.Find("td:nth-child(3)").Text()
			period.Precipitation = row.Find("td:nth-child(4)").Text()
			period.Wind = strings.TrimSpace(row.Find("td:nth-child(5)").Text())

			day.Periods = append(day.Periods, period)
		})

		result = append(result, day)
	})

	return result
}

func getDocument(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	fmt.Println("status", url, res.StatusCode)
	if res.StatusCode != 200 {
		return nil, errors.New("failed to fetch")
	}

	return goquery.NewDocumentFromReader(res.Body)
}

func getWeatherReport(city, country string) ([]WeatherDay, error) {
	doc, err := getDocument("https://www.yr.no/soek/soek.aspx?spr=eng&&sted=" + city + "&land=" + country)
	if err != nil {
		return nil, err
	}

	href, exists := doc.Find("table.yr-table td a").Attr("href")
	if !exists {
		// assume that we got redirected to the weather page of the city
		return weatherReportFromDocument(doc), nil
	}

	// otherwise follow the first link on the search results page
	doc, err = getDocument("https://www.yr.no" + href)
	if err != nil {
		return nil, err
	}
	return weatherReportFromDocument(doc), nil
}

func getLocation() (Location, error) {
	res, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return Location{}, err
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Location{}, err
	}

	var loc Location
	json.Unmarshal(bytes, &loc)
	return loc, nil
}

type WeatherDay struct {
	Title   string
	Periods []WeatherPeriod
}

type WeatherPeriod struct {
	Time          string
	Forecast      string
	Temp          string
	Precipitation string
	Wind          string
}

type Location struct {
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
}
