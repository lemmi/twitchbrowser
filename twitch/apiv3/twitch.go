package apiv3

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/lemmi/closer"
	"github.com/lemmi/twitchbrowser/twitch"
	"github.com/pkg/errors"
)

var _ twitch.API = apiv3{}

type apiv3 struct {
	endpoint url.URL
	client   *http.Client
	headers  http.Header
}

// New returns a twitch.API that implements the twitch v3 api
func New(clientID string) twitch.API {
	return apiv3{
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

func (api apiv3) GetChannels(names []string) (twitch.Channels, error) {
	data := make(url.Values)
	data.Add("channel", strings.Join(names, ","))
	const limit = 100
	data.Set("limit", strconv.Itoa(limit))
	tr := api.newTwitchRequest("streams", data)
	resp := twitchresponse{}

	var chans twitch.Channels

	for tr.scan(&resp) {
		for _, stream := range resp.Streams {
			chans = append(chans, twitch.Channel{
				Streamer:    stream.Name,
				Description: stream.Status,
				Game:        stream.Game,
				Viewers:     stream.Viewers,
			})
		}
		if len(resp.Streams) < limit {
			break
		}
	}

	return chans, tr.err
}

func (api apiv3) doRequest(url string) (*http.Response, error) {
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
	api     apiv3
	nexturl string
	err     error
}

func (api apiv3) newTwitchRequest(method string, data ...url.Values) *twitchRequest {
	url := api.endpoint
	url.Path += method
	if len(data) == 1 {
		url.RawQuery = data[0].Encode()
	}
	return &twitchRequest{api: api, nexturl: url.String()}
}

func (tr *twitchRequest) scan(tresp *twitchresponse) bool {
	if tr.nexturl == "" {
		return false
	}

	var resp *http.Response
	resp, tr.err = tr.api.doRequest(tr.nexturl)
	if tr.err != nil {
		return false
	}
	defer closer.Do(resp.Body)

	tr.err = json.NewDecoder(resp.Body).Decode(tresp)
	if tr.err != nil {
		return false
	}

	tr.nexturl = tresp.Links.Next
	if tr.nexturl == "" {
		tr.err = nil
		return false
	}

	return true
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

type twitchresponse struct {
	Streams []Stream `json:"streams"`
	Links   struct {
		Next string `json:"next"`
	} `json:"_links"`
}
