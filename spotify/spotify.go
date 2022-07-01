package spotify

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"time"

	"github.com/d-andrii/spotify-playback/rand"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const RedirectUrl = "http://localhost:4613/callback"

var (
	ClientId     string
	ClientSecret string
)

var (
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(RedirectUrl),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
		),
		spotifyauth.WithClientID(ClientId),
		spotifyauth.WithClientSecret(ClientSecret),
	)
	state = rand.RandString(10)
)

type TimeRange struct {
	StartTime string
	EndTime   string
}

type Client struct {
	device string
	client *spotify.Client
	ch     chan bool
	time   TimeRange
	cron   *cron.Cron
}

func New() Client {
	c := cron.New()
	c.Start()

	sc := Client{
		client: nil,
		ch:     make(chan bool, 1),
		cron:   c,
	}

	if err := sc.GetFromConfig(); err != nil {
		log.Println(err)
		sentry.CaptureException(err)
	}

	if err := sc.SetSchedulerTime("10:00", "22:00"); err != nil {
		log.Println(err)
		sentry.CaptureException(err)
	}

	return sc
}

func (sc *Client) HandleCallback(r *http.Request) error {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		return err
	}
	if st := r.FormValue("state"); st != state {
		return fmt.Errorf("state mismatch: %s != %s", st, state)
	}

	sc.client = spotify.New(auth.Client(r.Context(), tok))
	sc.ch <- true
	close(sc.ch)

	return nil
}

func (sc *Client) GetAuthUrl() string {
	return auth.AuthURL(state)
}

func (sc *Client) GetClient() *spotify.Client {
	if sc.client == nil {
		<-sc.ch
	}
	return sc.client
}

func (sc *Client) GetDevice(ctx context.Context) (string, error) {
	if sc.device == "" {
		p, err := sc.client.PlayerState(ctx)
		if err != nil {
			return "", err
		}

		sc.device = p.Device.ID.String()
	}

	return sc.device, nil
}

func (sc *Client) SetDevice(device string) {
	sc.device = device
	if err := sc.SaveConfig(); err != nil {
		log.Println(err)
	}
}

func (sc *Client) SetPlayerStatus(active bool) error {
	ctx := context.Background()

	ps, err := sc.client.PlayerState(ctx)
	if err != nil {
		return err
	}

	id := spotify.ID(sc.device)
	opts := spotify.PlayOptions{DeviceID: &id}

	if active && !ps.Playing {
		if err := sc.client.PlayOpt(ctx, &opts); err != nil {
			return err
		}
	} else if !active && ps.Playing {
		if err := sc.client.PauseOpt(ctx, &opts); err != nil {
			return err
		}
	}

	if err := sc.SaveConfig(); err != nil {
		return err
	}

	return nil
}

func (sc *Client) GetSchedulerTime() TimeRange {
	return sc.time
}

func (sc *Client) SetSchedulerTime(startTime string, endTime string) error {
	sc.time = TimeRange{startTime, endTime}
	for _, e := range sc.cron.Entries() {
		sc.cron.Remove(e.ID)
	}

	st, err := time.Parse("15:04", startTime)
	if err != nil {
		return err
	}
	et, err := time.Parse("15:04", endTime)
	if err != nil {
		return err
	}

	if _, err = sc.cron.AddFunc(fmt.Sprintf("%d %d * * *", st.Minute(), st.Hour()), func() {
		if err := sc.SetPlayerStatus(true); err != nil {
			log.Println(err)
			sentry.CaptureException(err)
		}
	}); err != nil {
		return err
	}

	if _, err = sc.cron.AddFunc(fmt.Sprintf("%d %d * * *", et.Minute(), et.Hour()), func() {
		if err := sc.SetPlayerStatus(false); err != nil {
			log.Println(err)
			sentry.CaptureException(err)
		}
	}); err != nil {
		return err
	}

	if err := sc.SaveConfig(); err != nil {
		log.Println(err)
		sentry.CaptureException(err)
	}

	return nil
}
