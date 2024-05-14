package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func logError(err error) {
	log.Fatal("Ups! an error ocurred when trying to get the weather information.", err)
}

type Weather struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
	Forecast struct {
		ForecastDay []struct {
			Hour []struct {
				TimeEpoch int64   `json:"time_epoch"`
				TempC     float64 `json:"temp_c"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
				ChanceOfRain float64 `json:"chance_of_rain"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logError(err)
	}

	apiKey := os.Getenv("API_KEY")

	// get the user location
	userLocation := ""
	fmt.Println("Please insert the location (Default: Lima)")
	fmt.Scanf("%s\n", &userLocation)
	if userLocation == "" {
		userLocation = "Lima"
	}
	endpoint := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%v&q=%v&days=1&aqi=no&alerts=no", apiKey, userLocation)

	res, err := http.Get(endpoint)
	if err != nil {
		logError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		logError(errors.New("invalid code"))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logError(err)
	}

	var weather Weather
	err = json.Unmarshal(body, &weather)

	if err != nil {
		logError(err)
	}

	location, current, hours := weather.Location, weather.Current, weather.Forecast.ForecastDay[0].Hour

	fmt.Printf("%s, %s: %.0fC, %s\n",
		location.Name,
		location.Country,
		current.TempC,
		current.Condition.Text,
	)

	for _, hour := range hours {
		date := time.Unix(hour.TimeEpoch, 0)

		// if date.Before(time.Now()) {
		// 	continue
		// }

		message := fmt.Sprintf("%s - %.0fC, %.0f%%, %s\n",
			date.Format("15:04"),
			hour.TempC,
			hour.ChanceOfRain,
			hour.Condition.Text,
		)
		if hour.ChanceOfRain < 40 {
			fmt.Print(message)
		} else {
			color.Red(message)
		}
	}
}
