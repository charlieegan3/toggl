package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type entry struct {
	Description string `json:"description"`
	Duration    int64  `json:"duration"`
	Start       string `json:"start"`
}

func getEntries(day string) (string, error) {
	req, err := http.NewRequest("GET", "https://www.toggl.com/api/v8/time_entries", nil)
	if err != nil {
		return "", fmt.Errorf("Failed to create request: %v", err)
	}
	req.SetBasicAuth(os.ExpandEnv("$TOKEN"), "api_token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to get entries: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}

	var entries []entry
	err = json.Unmarshal(body, &entries)
	if err != nil {
		return "", fmt.Errorf("Failed to parse json body: %v", err)
	}

	entriesForDate := make(map[string]int64)
	for _, v := range entries {
		if strings.Contains(v.Start, day) {
			entriesForDate[v.Description] += v.Duration
		}
	}

	var text string
	for k, v := range entriesForDate {
		if v > 0 {
			text += fmt.Sprintf("- [%.2f] %s\n", float64(v)/60.0/60.0, k)
		}
	}

	return text, nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		layout := "2006-01-02"

		// set a default date
		yesterday := time.Now().AddDate(0, 0, -1)
		if yesterday.Format("Monday") == "Sunday" {
			yesterday = time.Now().AddDate(0, 0, -3)
		}
		date := yesterday.Format(layout)

		// attempt to use date passed as a param
		dates, ok := r.URL.Query()["date"]
		if ok && len(dates) > 0 {
			date = dates[0]
		}

		// parse date as a time to allow getting of day
		day, err := time.Parse(layout, date)
		if err != nil {
			fmt.Fprintf(w, fmt.Sprintf("failed to parse date: %v", err))
		}

		// compute the next day for the second half of the standup
		nextDay := day.AddDate(0, 0, 1)

		// fetch the entries
		entries, err := getEntries(date)
		if err != nil {
			fmt.Fprintf(w, fmt.Sprintf("failed to get entries: %v", err))
		}

		standupText := fmt.Sprintf(
			"%s\n%s\n%s\n-",
			day.Format("Monday"),
			entries,
			nextDay.Format("Monday"),
		)

		fmt.Fprintf(w, standupText)
	})

	http.ListenAndServe(":8080", nil)
}
