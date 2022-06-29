package spotify

import (
	"encoding/json"
	"github.com/d-andrii/spotify-playback/helper"
	"os"
)

type Config struct {
	Device string
	Time   TimeRange
}

func (sc *Client) GetFromConfig() error {
	var c Config
	d, err := os.ReadFile("config.json")
	if err != nil {
		return helper.If(err == os.ErrNotExist, nil, err)
	}

	if err := json.Unmarshal(d, &c); err != nil {
		return err
	}

	sc.device = c.Device
	sc.time = c.Time

	return nil
}

func (sc *Client) SaveConfig() error {
	d, err := json.Marshal(Config{
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
