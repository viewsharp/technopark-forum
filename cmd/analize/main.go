package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
)

type Row struct {
	Duration float64 `json:"duration"`
	URI      string  `json:"uri"`
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run() error {
	logs, err := os.Open("logs.jsonl")
	if err != nil {
		return err
	}
	defer logs.Close()

	var rows []Row
	decoder := json.NewDecoder(logs)
	for decoder.More() {
		var row Row
		err = decoder.Decode(&row)
		if err != nil {
			return err
		}
		rows = append(rows, row)
	}

	durationByURI := make(map[string]float64)
	for _, row := range rows {
		parts := strings.Split(row.URI, "/")
		parts[3] = "-"
		uri := strings.Join(parts, "/")

		url, err := url.Parse(uri)
		if err != nil {
			return err
		}

		query := url.Query()
		if query.Has("limit") {
			query.Set("limit", "0")
		}
		if query.Has("desc") {
			query.Set("desc", "0")
		}
		if query.Has("since") {
			query.Set("since", "0")
		}

		durationByURI[url.Path+"?"+query.Encode()] += row.Duration
	}

	var durations []Row
	for uri, duration := range durationByURI {
		durations = append(durations, Row{duration, uri})
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i].Duration < durations[j].Duration
	})

	for _, row := range durations {
		fmt.Println(row.URI, row.Duration)
	}

	return nil
}
