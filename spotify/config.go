package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/d-andrii/spotify-playback/helper"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"os"
)

type Config struct {
	Token  *oauth2.Token
	Device string
	Time   TimeRange
}

func (sc *Client) GetFromConfig() error {
	var c Config
	d, err := os.ReadFile("config.json")
	if err != nil {
		return helper.If(errors.Is(err, os.ErrNotExist), nil, err)
	}

	if err := json.Unmarshal(d, &c); err != nil {
		return err
	}

	sc.device = c.Device
	sc.time = c.Time
	if c.Token != nil {
		sc.client = spotify.New(auth.Client(context.Background(), c.Token))
		sc.ch <- true
		close(sc.ch)
	}

	return nil
}

func (sc *Client) SaveConfig() error {
	var t *oauth2.Token
	var err error
	if sc.client != nil {
		if t, err = sc.client.Token(); err != nil {
			return err
		}
	}

	d, err := json.Marshal(Config{
		Token:  t,
		Device: sc.device,
		Time:   sc.time,
	})
	if err != nil {
		return err
	}

	if err := os.WriteFile("config.json", d, 0644); err != nil {
		return err
	}

	return nil
}
