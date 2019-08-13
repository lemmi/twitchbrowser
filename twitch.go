package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type twitchAPI struct {
	endpoint url.URL
	client   *http.Client
	headers  http.Header
}

func newtwitchAPI(clientID string) twitchAPI {
	return twitchAPI{
		endpoint: url.URL{
			Scheme: "https",
			Host:   "api.twitch.tv",
			Path:   "/kraken/",
		},
		client: &http.Client{},
		headers: http.Header{
			"Client-ID": []string{clientID},
			"Accept":    []string{"application/vnd.twitchtv.v3+json"},
		},
	}
}

func (api twitchAPI) RawTwitchRequest(url string, clientID string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for k, vs := range api.headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	req.Close = true

	resp, err := api.client.Do(req)
	if resp != nil && resp.StatusCode != 200 {
		err = errors.New(resp.Status)
	}
	return resp, err
}

type twitchRequest struct {
	api      twitchAPI
	clientID string
	nexturl  string
	err      error
}

func (api twitchAPI) NewTwitchRequest(method string, data ...url.Values) *twitchRequest {
	url := api.endpoint
	url.Path += method
	if len(data) == 1 {
		url.RawQuery = data[0].Encode()
	}
	return &twitchRequest{api: api, nexturl: url.String()}
}

func (tr *twitchRequest) Scan(s nexter) bool {
	if tr.nexturl == "" {
		return false
	}

	var resp *http.Response
	resp, tr.err = tr.api.RawTwitchRequest(tr.nexturl, tr.clientID)
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

func (tr *twitchRequest) Err() error {
	return tr.err
}

type nexter interface {
	Next() string
}

// Links holds info for pagination
type Links struct {
	Links struct {
		Next string `json:"next"`
	} `json:"_links"`
}

// Next returns the link for the next page
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

// Channel holds channel info
type Channel struct {
	Name   string
	Status string
}

// Stream holds stream info
type Stream struct {
	Game    string
	Viewers int
	Channel `json:"channel"`
}

func channelNames(names []string) url.Values {
	return url.Values{
		"channel": []string{strings.Join(names, ",")},
	}
}

func (api twitchAPI) GetChannels(data url.Values) (Chans, error) {
	type twitchresponse struct {
		Streams []Stream `json:"streams"`
		Links
	}

	const limit = 100
	data["limit"] = []string{strconv.Itoa(limit)}
	tr := api.NewTwitchRequest("streams", data)
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

func (api twitchAPI) GetChannelsFunc(names []string) func() (Chans, error) {
	return func() (Chans, error) {
		return api.GetChannels(channelNames(names))
	}
}
