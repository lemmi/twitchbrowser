package main

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	httpclient = http.Client{
		Transport: &http.Transport{
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	twitchApiUrl = url.URL{
		Scheme: "https",
		Host:   "api.twitch.tv",
		Path:   "/kraken/",
	}
)

func TwitchRequest(method string, data url.Values) (resp *http.Response, err error) {
	url := twitchApiUrl
	url.Path += method
	url.RawQuery = data.Encode()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("Client-ID", "browser")
	req.Header.Set("Accept", "application/vnd.twitchtv.v2+json")
	req.Close = true

	resp, err = httpclient.Do(req)
	return
}

type twitchstream struct {
	streamer    string
	description string
	game        string
	viewers     int
}

func (t twitchstream) Streamer() string {
	return t.streamer
}
func (t twitchstream) Description() string {
	return t.description
}
func (t twitchstream) Game() string {
	return t.game
}
func (t twitchstream) Viewers() int {
	return t.viewers
}

func LoadFavs() (names []string) {
	user, err := user.Current()
	if err != nil {
		return
	}

	path := filepath.Join(user.HomeDir, ".config", "twitchbrowser", "favorites.conf")
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if len(name) > 0 {
			names = append(names, name)
		}
	}
	return
}

func GetFavChannels() (chans Chans, err error) {
	return GetChannels(ChannelNames(LoadFavs()))
}

func ChannelNames(names []string) url.Values {
	return url.Values{
		"channel": []string{strings.Join(names, ",")},
	}
}

func GameName(name string) url.Values {
	return url.Values{
		"game": []string{name},
	}
}

func GetChannels(data url.Values) (chans Chans, err error) {
	const limit = 100
	data["limit"] = []string{strconv.Itoa(limit)}

	resp, err := TwitchRequest("streams", data)
	if err != nil {
		return
	}

	type inner struct {
		Name   string
		Status string
	}
	type stream struct {
		Game    string
		Viewers int
		inner   `json:"Channel"`
	}

	streams := []stream{}

	err = json.NewDecoder(resp.Body).Decode(&struct{ Streams *[]stream }{&streams})
	resp.Body.Close()

	for _, stream := range streams {
		chans = append(chans, &twitchstream{
			streamer:    stream.Name,
			description: stream.Status,
			game:        stream.Game,
			viewers:     stream.Viewers,
		})
	}
	return
}

func GetGameFunc(name string) func() (Chans, error) {
	return func() (Chans, error) {
		return GetChannels(GameName(name))
	}
}
func GetChannelsFunc(names []string) func() (Chans, error) {
	return func() (Chans, error) {
		return GetChannels(ChannelNames(names))
	}
}
