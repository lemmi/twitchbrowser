package apinew

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/lemmi/closer"
	"github.com/lemmi/twitchbrowser/twitch"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"
	oauth2twitch "golang.org/x/oauth2/twitch"
)

var (
	endpoint = url.URL{
		Scheme: "https",
		Host:   "api.twitch.tv",
		Path:   "/helix/",
	}
)

var _ twitch.API = apinew{}

type apinew struct {
	headers http.Header
	client  *http.Client
	*sync.Mutex
	game map[string]string
}

func New(clientID string, secret string) twitch.API {
	oauth2Config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		TokenURL:     oauth2twitch.Endpoint.TokenURL,
	}
	return apinew{
		headers: http.Header{
			"Client-ID": []string{clientID},
		},
		client: oauth2Config.Client(context.Background()),
		Mutex:  new(sync.Mutex),
		game:   make(map[string]string),
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func nextChunk(s []string, n int) ([]string, []string) {
	split := min(len(s), n)
	return s[:split], s[split:]
}

func (api apinew) GetChannels(names []string) (twitch.Channels, error) {
	var ret twitch.Channels
	for len(names) > 0 {
		var cur []string
		cur, names = nextChunk(names, 100)
		chans, err := api.getChannels(cur)
		if err != nil {
			return ret, err
		}
		ret = append(ret, chans...)
	}

	return ret, api.mapGameID(ret)
}

func (api apinew) getChannels(names []string) (twitch.Channels, error) {
	data := make(url.Values)

	for _, name := range names {
		data.Add("user_login", name)
	}

	var resp getStreamsResp
	err := api.doRequest("streams", data, &resp)
	if err != nil {
		return nil, err
	}

	var ret twitch.Channels
	for _, d := range resp.Data {
		ret = append(ret, twitch.Channel{
			Streamer:    d.UserName,
			Description: d.Title,
			Game:        d.GameID,
			Viewers:     d.ViewerCount,
		})
	}

	return ret, nil
}
func (api apinew) mapGameID(chans twitch.Channels) error {
	api.Lock()
	defer api.Unlock()

	missing := make(map[string]struct{})

	for _, c := range chans {
		if _, ok := api.game[c.Game]; !ok {
			missing[c.Game] = struct{}{}
		}
	}

	ids := make([]string, 0, len(missing))
	for id := range missing {
		ids = append(ids, id)
	}

	mapping, err := api.GetGameNames(ids)
	if err != nil {
		return err
	}

	for id, name := range mapping {
		api.game[id] = name
	}

	for i, c := range chans {
		chans[i].Game = api.game[c.Game]
	}

	return nil
}

func (api apinew) GetGameNames(ids []string) (map[string]string, error) {
	var ret map[string]string
	for len(ids) > 0 {
		var cur []string
		cur, ids = nextChunk(ids, 100)
		mappings, err := api.getGameNames(cur)
		if err != nil {
			return ret, err
		}
		if ret == nil {
			ret = mappings
		} else {
			for k, v := range mappings {
				ret[k] = v
			}
		}
	}

	return ret, nil
}
func (api apinew) getGameNames(ids []string) (map[string]string, error) {
	data := make(url.Values)
	for _, id := range ids {
		data.Add("id", id)
	}
	var resp getGameResp
	err := api.doRequest("games", data, &resp)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]string, len(resp.Data))
	for _, d := range resp.Data {
		ret[d.ID] = d.Name
	}
	return ret, nil
}

type getGameResp struct {
	Data []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		BoxArtURL string `json:"box_art_url"`
	} `json:"data"`
}

func (api apinew) doRequest(method string, data url.Values, v interface{}) error {
	u := endpoint
	u.Path = path.Join(u.Path, method)
	query := data
	if query == nil {
		query = make(url.Values)
	}
	query.Set("first", "100")
	u.RawQuery = query.Encode()

	//log.Println("Requesting", api.headers, u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return errors.Wrapf(err, "failed buidling request to %q", u.String())
	}
	req.Header = api.headers
	req.Close = true

	resp, err := api.client.Do(req)
	switch {
	case resp != nil && resp.StatusCode != 200:
		return errors.New(resp.Status)
	case err != nil:
		return errors.Wrapf(err, "error in response to %q", u.String())
	}
	defer closer.Do(resp.Body)

	return errors.Wrap(
		json.NewDecoder(resp.Body).Decode(v),
		"error decoding json",
	)
}

type getStreamsResp struct {
	Data       []data     `json:"data"`
	Pagination pagination `json:"pagination"`
}
type data struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	CommunityIds []string  `json:"community_ids"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
}
type pagination struct {
	Cursor string `json:"cursor"`
}
