package main

import (
	"encoding/json"
	"html"
	"net/http"
)

const (
	srlAPIURL = "http://api.speedrunslive.com/frontend/streams"
)

func getSRLNames() (twitchnames []string) {
	resp, err := http.Get(srlAPIURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	t := struct {
		Source struct {
			Channels []struct{ Name string }
		} `json:"_source"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return
	}
	twitchnames = make([]string, len(t.Source.Channels))
	for i, p := range t.Source.Channels {
		twitchnames[i] = html.UnescapeString(p.Name)
	}
	return twitchnames
}
