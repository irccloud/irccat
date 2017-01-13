package httplistener

import (
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

	mux.HandleFunc("/send", hl.genericHandler)

	if viper.IsSet("http.listeners.grafana") {
		mux.HandleFunc("/grafana", hl.grafanaAlertHandler)
	}

	hl.http.Handler = mux
	log.Infof("Listening for HTTP requests on %s", viper.GetString("http.listen"))
	if viper.GetBool("http.ssl") {
		go hl.http.ListenAndServeTLS(viper.GetString("http.ssl_cert"), viper.GetString("http.ssl_key"))
	} else {
		go hl.http.ListenAndServe()
	}
	return &hl, nil
}
