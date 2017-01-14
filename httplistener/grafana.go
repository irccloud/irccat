package httplistener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
)

type grafanaMatch struct {
	Metric string
	Value  float32
}

type grafanaAlert struct {
	Title       string
	RuleName    string
	RuleUrl     string
	State       string
	ImageUrl    string
	Message     string
	EvalMatches []grafanaMatch
}

func (hl *HTTPListener) grafanaAlertHandler(w http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.NotFound(w, request)
		return
	}

	var alert grafanaAlert
	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	json.Unmarshal(buf.Bytes(), &alert)
	msg := fmt.Sprintf("[Grafana] [%s] %s: %s.", alert.State, alert.RuleName, alert.Message)
	for _, match := range alert.EvalMatches {
		msg += fmt.Sprintf(" %s:%f", match.Metric, match.Value)
	}
	msg += " " + alert.RuleUrl

	log.Infof("%s [%s] Grafana alert", request.RemoteAddr, viper.GetString("http.listeners.grafana"))
	hl.irc.Privmsgf(viper.GetString("http.listeners.grafana"), msg)
}
