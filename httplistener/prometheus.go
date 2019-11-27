package httplistener

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/viper"
	"net/http"
)

type HookMessage struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []Alert           `json:"alerts"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt,omitempty"`
	EndsAt       string            `json:"endsAt,omitempty"`
	GeneratorURL string            `json:"generatorURL"`
}

func (hl *HTTPListener) prometheusHandler(w http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.NotFound(w, request)
		return
	}

	var message HookMessage
	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	json.Unmarshal(buf.Bytes(), &message)
	log.Infof("%s [%s] Prometheus alert", request.RemoteAddr, viper.GetString("http.listeners.prometheus"))

	msgs, err := hl.renderTemplate("prometheus.alert", message)
	if err != nil {
		log.Errorf("Unable to render template: %s", err)
		return
	}

	channel := viper.GetString("http.listeners.prometheus")
	for _, msg := range msgs {
		hl.irc.Privmsg(channel, msg)
	}
}
