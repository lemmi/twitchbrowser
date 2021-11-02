package main

import (
	"encoding/json"
	"html"
	"net/http"

	"github.com/lemmi/closer"
)

const (
	srlAPIURL = "https://www.speedrunslive.com/api/liveStreams?pageNumber=1&search=&pageSize=10000"
)

func getSRLNames() (twitchnames []string) {
	resp, err := http.Get(srlAPIURL)
	if err != nil {
		return
	}
	defer closer.Do(resp.Body)

	t := struct {
		Data struct {
			Livestreams struct {
				Data []struct{ Name string }
			}
		}
	}{}
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return
	}
	twitchnames = make([]string, len(t.Data.Livestreams.Data))
	for i, p := range t.Data.Livestreams.Data {
		twitchnames[i] = html.UnescapeString(p.Name)
	}
	return twitchnames
}
