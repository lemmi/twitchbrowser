package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
)

const (
	srlApiUrl = "http://api.speedrunslive.com/test/team"
)

type srlChannel struct {
	Display_name    string
	Current_viewers json.Number
	Title           string
	Name            string
	Meta_game       string
}

func (c srlChannel) Streamer() string {
	return c.Name
}
func (c srlChannel) Description() string {
	return c.Title
}
func (c srlChannel) Game() string {
	return c.Meta_game
}
func (c srlChannel) Viewers() int {
	n, err := c.Current_viewers.Int64()
	if err != nil {
		fmt.Println(err)
		n = 0
	}
	return int(n)
}

func (c srlChannel) Unescape() srlChannel {
	c.Display_name = html.UnescapeString(c.Display_name)
	c.Title = html.UnescapeString(c.Title)
	c.Name = html.UnescapeString(c.Name)
	c.Meta_game = html.UnescapeString(c.Meta_game)

	return c
}

func GetSRLChannels() (chans Chans, err error) {
	resp, err := http.Get(srlApiUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	t := struct {
		Channels []struct{ Channel srlChannel }
	}{}
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return
	}
	chans = make(Chans, len(t.Channels))
	for i, p := range t.Channels {
		unescaped := p.Channel.Unescape()
		chans[i] = &unescaped
	}
	return
}
