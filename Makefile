include .env

OUT_DIR=bin
OUT_APP_NAME=SpotifyPlayback

GO_VERSION=1.18.3

windows: GOOS=windows
windows: GOARCH=amd64
windows:
	GOOS=${GOOS} GOARCH=${GOARCH} go${GO_VERSION} build -ldflags "-X 'github.com/d-andrii/spotify-playback/spotify.ClientId=${CLIENT_ID}' -X 'github.com/d-andrii/spotify-playback/spotify.ClientSecret=${CLIENT_SECRET}' -H windowsgui" -o ${OUT_DIR}/${OUT_APP_NAME}-${GOOS}-${GOARCH}.exe

macos: GOOS=darwin
macos: GOARCH=arm64
macos:
	GOOS=${GOOS} GOARCH=${GOARCH} go${GO_VERSION} build -ldflags "-X 'github.com/d-andrii/spotify-playback/spotify.ClientId=${CLIENT_ID}' -X 'github.com/d-andrii/spotify-playback/spotify.ClientSecret=${CLIENT_SECRET}'" -o ${OUT_DIR}/${OUT_APP_NAME}-${GOOS}-${GOARCH}

all: macos windows