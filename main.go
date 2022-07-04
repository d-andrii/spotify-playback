package main

import (
	"github.com/getlantern/systray"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/browser"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/d-andrii/spotify-playback/helper"
	"github.com/d-andrii/spotify-playback/spotify"
	spotifySource "github.com/zmb3/spotify/v2"

	_ "embed"
)

type IndexData struct {
	Devices       []spotifySource.PlayerDevice
	CurrentDevice string
	CurrentState  string
	StartTime     string
	EndTime       string
}

//go:embed index.gohtml
var IndexTemplate string

var spotifyClient = spotify.New()

func check(action string, err error) {
	log.Printf("trying to %s\n", action)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatalf("failed to %s: %v\n", action, err)
	}
}

func status(w http.ResponseWriter, action string, code int, err error) bool {
	log.Printf("trying to %s\n", action)
	if err != nil {
		log.Printf("failed to %s: %v\n", action, err)
		sentry.CaptureException(err)
		w.WriteHeader(code)
		_, _ = io.WriteString(w, err.Error())

		return true
	}

	return false
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	if status(w, "handle callback", http.StatusBadRequest, spotifyClient.HandleCallback(r)) {
		return
	}

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	if status(w, "parse form", http.StatusBadRequest, r.ParseForm()) {
		return
	}

	spotifyClient.SetDevice(r.FormValue("device"))
	if status(w, "set player status", http.StatusInternalServerError, spotifyClient.SetPlayerStatus(r.FormValue("status") == "play")) {
		return
	}
	if status(w, "set scheduler time", http.StatusInternalServerError, spotifyClient.SetSchedulerTime(r.FormValue("startTime"), r.FormValue("endTime"))) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	logFile, err := os.OpenFile("./SpotifyPlayback.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	log.Println("Main")
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://9fabaa56be03478db940886f40668c6a@o1304179.ingest.sentry.io/6544256",
		TracesSampleRate: 1.0,
		AttachStacktrace: true,
	}); err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	log.Println("Starting tray")
	systray.Run(onReady, onExit)
}

func setupTray() {
	systray.SetIcon(helper.Icon)
	mStgs := systray.AddMenuItem("Налаштування", "Відкрити налаштування у браузері")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Вийти", "Закрити застосунок")

	log.Println("Setting up tray handlers")

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-mStgs.ClickedCh:
				if err := browser.OpenURL("http://localhost:4613"); err != nil {
					log.Println(err)
					sentry.CaptureException(err)
				}
			}
		}
	}()
}

func onReady() {
	setupTray()

	t, err := template.New("main").Parse(IndexTemplate)
	check("parse template", err)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		client := spotifyClient.GetClient()

		devices, err := client.PlayerDevices(r.Context())
		status(w, "get player devices", http.StatusInternalServerError, err)

		device, err := spotifyClient.GetDevice(r.Context())
		status(w, "get current device", http.StatusInternalServerError, err)

		p, err := client.PlayerState(r.Context())
		status(w, "get current player state", http.StatusInternalServerError, err)

		st := spotifyClient.GetSchedulerTime()

		check("execute template", t.Execute(w, IndexData{
			Devices:       devices,
			CurrentDevice: device,
			CurrentState:  helper.If(p.Playing, "play", "pause"),
			StartTime:     st.StartTime,
			EndTime:       st.EndTime,
		}))
	})
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/save", handleSave)

	url := spotifyClient.GetAuthUrl()

	log.Println(url)

	if err := browser.OpenURL(url); err != nil {
		log.Println(err)
		sentry.CaptureException(err)
	}

	check("start http server", http.ListenAndServe(":4613", nil))
}

func onExit() {
	sentry.Flush(2 * time.Second)
}
