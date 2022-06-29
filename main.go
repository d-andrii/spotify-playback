package main

import (
	"github.com/d-andrii/spotify-playback/helper"
	"github.com/d-andrii/spotify-playback/spotify"
	spotifySource "github.com/zmb3/spotify/v2"
	"html/template"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"

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
	if err != nil {
		log.Fatalf("failed to %s: %v\n", action, err)
	}
}

func status(w http.ResponseWriter, action string, code int, err error) bool {
	if err != nil {
		log.Printf("failed to %s: %v\n", action, err)
		_, _ = io.WriteString(w, err.Error())
		w.WriteHeader(code)

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
	status(w, "set player status", http.StatusInternalServerError, spotifyClient.SetPlayerStatus(r.Context(), r.FormValue("status") == "play"))
	status(w, "set scheduler time", http.StatusInternalServerError, spotifyClient.SetSchedulerTime(r.FormValue("startTime"), r.FormValue("endTime")))

	w.WriteHeader(http.StatusNoContent)
}

func main() {
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

	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	}

	check("start http server", http.ListenAndServe(":4613", nil))
}
