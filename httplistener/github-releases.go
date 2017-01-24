package httplistener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
)

type githubRepository struct {
	Name        string
	Description string
}

type githubRelease struct {
	Html_url    string
	Tag_name    string
	Name        string
}

type githubPayload struct {
	Action      string
	Release     githubRelease
	Repository  githubRepository
}

func (hl *HTTPListener) githubReleasesHandler(w http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.NotFound(w, request)
		return
	}

	var payload githubPayload
	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	json.Unmarshal(buf.Bytes(), &payload)
	
	if payload.Action == "published" {
		name := payload.Repository.Description
		if name == "" {
			name = payload.Repository.Name
		}
		release := payload.Release.Name
		if release == "" {
			release = payload.Release.Tag_name
		}
		
		msg := fmt.Sprintf("[\x02Release\x0f] \x02%s\x0f version \x02%s\x0f has been published: %s", name, release, payload.Release.Html_url)

		channelKey := fmt.Sprintf("http.listeners.github-releases.%s", payload.Repository.Name)
		channel := viper.getString(channelKey)
		log.Infof("%s [%s] GitHub Release", request.RemoteAddr, channel)
		hl.irc.Privmsgf(channel, msg)
	}
}
