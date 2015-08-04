package main

import (
	"bufio"
	"encoding/json"
	"errors"
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

func RawTwitchRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-ID", "browser")
	req.Header.Set("Accept", "application/vnd.twitchtv.v3+json")
	req.Close = true

	resp, err := httpclient.Do(req)
	if resp != nil && resp.StatusCode != 200 {
		err = errors.New(resp.Status)
	}
	return resp, err

}

type TwitchRequest struct {
	nexturl string
	err     error
}

func NewTwitchRequest(method string, data ...url.Values) *TwitchRequest {
	url := twitchApiUrl
	url.Path += method
	if len(data) == 1 {
		url.RawQuery = data[0].Encode()
	}
	return &TwitchRequest{nexturl: url.String()}
}

func (tr *TwitchRequest) Scan(s Nexter) bool {
	if tr.nexturl == "" {
		return false
	}

	var resp *http.Response
	resp, tr.err = RawTwitchRequest(tr.nexturl)
	if tr.err != nil {
		return false
	}
	defer resp.Body.Close()

	tr.err = json.NewDecoder(resp.Body).Decode(s)
	if tr.err != nil {
		return false
	}

	tr.nexturl = s.Next()
	if tr.nexturl == "" {
		tr.err = nil
		return false
	}

	return true
}

func (tr *TwitchRequest) Err() error {
	return tr.err
}

type Nexter interface {
	Next() string
}

type Links struct {
	Links struct {
		Next string `json:"next"`
	} `json:"_links"`
}

func (l Links) Next() string {
	return l.Links.Next
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

type Channel struct {
	Name   string
	Status string
}

type Stream struct {
	Game    string
	Viewers int
	Channel `json:"channel"`
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

func GetFollowChannels(user string) func() (Chans, error) {
	names, err := GetFollows(user)
	if err != nil {
		return func() (Chans, error) {
			return nil, err
		}
	}
	return func() (Chans, error) {
		return GetChannels(ChannelNames(names))
	}
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

func GetChannels(data url.Values) (Chans, error) {
	type twitchresponse struct {
		Streams []Stream `json:"streams"`
		Links
	}

	const limit = 100
	data["limit"] = []string{strconv.Itoa(limit)}
	tr := NewTwitchRequest("streams", data)
	resp := twitchresponse{}

	var chans Chans

	for tr.Scan(&resp) {
		for _, stream := range resp.Streams {
			chans = append(chans, &twitchstream{
				streamer:    stream.Name,
				description: stream.Status,
				game:        stream.Game,
				viewers:     stream.Viewers,
			})
		}
		if len(resp.Streams) < limit {
			break
		}
	}

	return chans, tr.Err()
}

func GetFollows(user string) ([]string, error) {
	tr := NewTwitchRequest("/users/" + user + "/follows/channels")

	type followsresp struct {
		Follows []struct {
			Channel Channel
		}
		Links
	}

	ret := []string{}
	resp := followsresp{}

	for tr.Scan(&resp) {
		if len(resp.Follows) == 0 {
			break
		}

		for _, follow := range resp.Follows {
			ret = append(ret, follow.Channel.Name)
		}
	}

	return ret, tr.Err()
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
