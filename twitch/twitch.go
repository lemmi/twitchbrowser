package twitch

import "strings"

// API provides read only access to the twitch api
type API interface {
	GetChannels(names []string) (Channels, error)
}

// GetChannelsFunc wraps the GetChannels call for asny scheduling
func GetChannelsFunc(api API, names []string) func() (Channels, error) {
	return func() (Channels, error) {
		return api.GetChannels(names)
	}
}

// Channel stores channel data
type Channel struct {
	Streamer    string
	Description string
	Game        string
	Viewers     int
}

// Channels is used to sort []Channel
type Channels []Channel

func (chans Channels) Len() int {
	return len(chans)
}

func (chans Channels) Less(i, j int) bool {
	Eq := func(a, b string) bool {
		return strings.ToLower(a) == strings.ToLower(b)
	}
	Less := func(a, b string) bool {
		return strings.ToLower(a) < strings.ToLower(b)
	}

	chi := chans[i]
	chj := chans[j]
	switch {
	case !Eq(chi.Game, chj.Game):
		return Less(chi.Game, chj.Game)
	case chi.Viewers != chj.Viewers:
		return chi.Viewers > chj.Viewers
	case !Eq(chi.Streamer, chj.Streamer):
		return Less(chi.Streamer, chj.Streamer)
	default:
		return Less(chi.Description, chj.Description)
	}
}

func (chans Channels) Swap(i, j int) {
	chans[i], chans[j] = chans[j], chans[i]
}
