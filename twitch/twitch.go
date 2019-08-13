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
	Less := func(a, b string) bool {
		return strings.ToLower(a) < strings.ToLower(b)
	}

	chi := chans[i]
	chj := chans[j]
	switch {
	case Less(chi.Game, chj.Game):
		return true
	case Less(chj.Game, chi.Game):
		return false
	case chi.Viewers > chj.Viewers:
		return true
	case chi.Viewers < chj.Viewers:
		return false
	case Less(chi.Streamer, chj.Streamer):
		return true
	case Less(chj.Streamer, chi.Streamer):
		return false
	case Less(chi.Description, chj.Description):
		return true
	case Less(chj.Description, chi.Description):
		return false
	default:
		return false
	}
}

func (chans Channels) Swap(i, j int) {
	chans[i], chans[j] = chans[j], chans[i]
}
