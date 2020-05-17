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
	var loc Location
	if len(os.Args) == 1 {
		var err error
		loc, err = getLocation()
		if err != nil {
			fmt.Println("Couldn't determine your location. Please provide the city where you are on the command line.", err)
			return
		}
	} else {
		loc = Location{City: os.Args[1]}
	}

	weather, err := getWeatherReport(loc.City, loc.CountryCode)
	if err != nil {
		fmt.Println("Error getting weather report:", err)
		fmt.Println("Try updating vejr")
		return
	}

	fmt.Println("Weather forecast for", weather.Title)
	fmt.Println()
	for _, day := range weather.Days {
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

func weatherReportFromDocument(doc *goquery.Document) WeatherReport {
	result := WeatherReport{}
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

		result.Days = append(result.Days, day)
	})

	result.Title = doc.Find("h1 span").Text()
	return result
}

func getDocument(url string) (*goquery.Document, bool, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, false, err
	}
	req.AddCookie(&http.Cookie{
		Name:  "brp",
		Value: "spr=eng",
	})

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, false, errors.New("failed to fetch")
	}

	r, err := goquery.NewDocumentFromReader(res.Body)
	return r, res.Request.URL.String() != url, err
}

func getWeatherReport(city, country string) (WeatherReport, error) {
	if city == "" {
		return WeatherReport{}, fmt.Errorf("Couldn't find city")
	}
	doc, redirected, err := getDocument("https://www.yr.no/soek/soek.aspx?spr=eng&sted=" + replaceSpaces(city) + "&land=" + country)
	if err != nil {
		return WeatherReport{}, err
	}

	href, exists := doc.Find("table.yr-table td a").Attr("href")
	if !exists {
		// assume that we got redirected to the weather page of the city
		if redirected {
			return weatherReportFromDocument(doc), nil
		}

		return getWeatherReport(removeLastWord(city), country)
	}

	// otherwise follow the first link on the search results page
	doc, _, err = getDocument("https://www.yr.no" + href)
	if err != nil {
		return WeatherReport{}, err
	}
	return weatherReportFromDocument(doc), nil
}

func replaceSpaces(str string) string {
	return strings.Replace(str, " ", "%20", -1)
}

func removeLastWord(str string) string {
	i := strings.LastIndex(str, " ")
	if i == -1 {
		return ""
	}
	return str[:i]
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
	err = json.Unmarshal(bytes, &loc)
	if err != nil {
		return Location{}, err
	}
	return loc, nil
}

type WeatherReport struct {
	Title string
	Days  []WeatherDay
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
