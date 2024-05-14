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

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

type model struct {
	// initial | success
	status    string
	textInput textinput.Model
	weather   Weather
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Lima"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	return model{
		status:    "initial",
		textInput: ti,
		err:       nil,
	}
}

type (
	errMsg error
)

func (m model) Init() tea.Cmd {
	err := godotenv.Load()
	if err != nil {
		logError(err)
	}
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		// if the user press enter, get the weather information
		case tea.KeyEnter:
			apiKey := os.Getenv("API_KEY")
			userLocation := m.textInput.Value()
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
			m.weather = weather
			m.status = "success"

			return m, nil
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	// update the input text value
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.status == "success" {
		location, current, hours := m.weather.Location, m.weather.Current, m.weather.Forecast.ForecastDay[0].Hour

		header := color.GreenString(
			fmt.Sprintf("%s, %s: %.0fC, %s",
				location.Name,
				location.Country,
				current.TempC,
				current.Condition.Text,
			),
		)

		var body string

		for _, hour := range hours {
			date := time.Unix(hour.TimeEpoch, 0)

			// if date.Before(time.Now()) {
			// 	continue
			// }

			message := fmt.Sprintf("%s - %.0fC, %.0f%%, %s",
				date.Format("15:04"),
				hour.TempC,
				hour.ChanceOfRain,
				hour.Condition.Text,
			)
			if hour.ChanceOfRain < 40 {
				body = fmt.Sprintf("%s\n%s", body, color.BlueString(message))
			} else {
				body = fmt.Sprintf("%s\n%s", body, color.RedString(message))
			}
		}

		return fmt.Sprintf("\n%s\n%s",
			header,
			body,
		)
	}

	return fmt.Sprintf(
		"What's your current location?\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
