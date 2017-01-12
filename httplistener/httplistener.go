package httplistener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"net/http"
)

var log = loggo.GetLogger("HTTPListener")

type HTTPListener struct {
	http http.Server
	irc  *irc.Connection
}

func New(irc *irc.Connection) (*HTTPListener, error) {
	hl := HTTPListener{}
	hl.irc = irc
	hl.http = http.Server{Addr: viper.GetString("http.listen")}

	mux := http.NewServeMux()
	if viper.IsSet("http.listeners.grafana") {
		mux.HandleFunc("/grafana", hl.grafanaAlertHandler)
	}

	hl.http.Handler = mux
	log.Infof("Listening for HTTP requests on %s", viper.GetString("http.listen"))
	go hl.http.ListenAndServe()
	return &hl, nil
}

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
	hl.irc.Privmsgf(viper.GetString("http.listeners.grafana"), msg)
}
