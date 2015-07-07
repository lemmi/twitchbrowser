package main

import (
	"strings"
)

type Channeler interface {
	Streamer() string
	Description() string
	Game() string
	Viewers() int
}

type Chans []Channeler

func (chans Chans) Len() int {
	return len(chans)
}

func (chans Chans) Less(i, j int) bool {
	Less := func(a, b string) bool {
		return strings.ToLower(a) < strings.ToLower(b)
	}

	chi := chans[i]
	chj := chans[j]
	switch {
	case Less(chi.Game(), chj.Game()):
		return true
	case Less(chj.Game(), chi.Game()):
		return false
	case chi.Viewers() > chj.Viewers():
		return true
	case chi.Viewers() < chj.Viewers():
		return false
	case Less(chi.Streamer(), chj.Streamer()):
		return true
	case Less(chj.Streamer(), chi.Streamer()):
		return false
	case Less(chi.Description(), chj.Description()):
		return true
	case Less(chj.Description(), chi.Description()):
		return false
	default:
		return false
	}
}

func (chans Chans) Swap(i, j int) {
	chans[i], chans[j] = chans[j], chans[i]
}
