package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/olekukonko/tablewriter"
)

func main() {
	city, country, err := getLocation()
	if err != nil {
		fmt.Println("Couldn't determine your location. Please provide the city where you are on the command line.")
		return
	}

	url, _ := searchYr(city, country)
	weather, _ := getWeatherPage(url)
	// weather, _ := getWeatherPage("https://www.yr.no/place/Denmark/Capital/Copenhagen/")
	fmt.Println(url)

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

func getWeatherPage(url string) ([]WeatherDay, error) {
	doc, err := getDocument(url)
	if err != nil {
		return nil, err
	}

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

	return result, nil
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

func getDocument(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New("failed to fetch")
	}

	return goquery.NewDocumentFromReader(res.Body)
}

func getLocation() (string, string, error) {
	res, err := http.Get("https://www.iplocation.net/")
	if err != nil {
		return "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}

	imgSrc, exists := doc.Find("table td img").First().Attr("src")
	if !exists {
		return "", "", fmt.Errorf("Couldn't find location")
	}

	re := regexp.MustCompile(`([a-z]{2})\.gif`)
	country := strings.ToUpper(re.FindStringSubmatch(imgSrc)[1])
	city := doc.Find("table td:nth-child(4)").First().Text()
	return city, country, nil
}

func searchYr(city, country string) (string, error) {
	doc, err := getDocument("https://www.yr.no/soek/soek.aspx?spr=eng&&sted=" + city + "&land=" + country)
	if err != nil {
		return "", err
	}

	href, exists := doc.Find("table.yr-table td a").Attr("href")
	if !exists {
		return "", fmt.Errorf("Couldn't find place in YR.")
	}
	return "https://www.yr.no" + href, nil
}
