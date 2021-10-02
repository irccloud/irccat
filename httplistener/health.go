package httplistener

import (
	"fmt"
	"net/http"
	"time"

	"github.com/irccloud/irccat/util"
	"github.com/spf13/viper"
)

var LagInterval = 10 * time.Second // for testing
var DefaultPeriod = 15 * time.Minute

// healthHandler returns non-2xx if the configured timeout has elapsed.
//
// This mainly exists to present an interface for supervisors relying on
// liveliness/readiness probes (e.g., for Kubernetes deployments). However, a
// conservative client could query it before sending a payload. See also
// /healthcheck in this same pkg.
func (hl *HTTPListener) healthHandler(
	w http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != "GET" {
		http.NotFound(w, request)
		return
	}
	healthFile := viper.GetString("irc.health_file")
	if healthFile == "" {
		http.NotFound(w, request)
		return
	}
	viper.SetDefault("irc.health_period", DefaultPeriod)

	freq := LagInterval + viper.GetDuration("irc.health_period")
	err := util.CheckTimestamp(healthFile, freq)
	if err != nil {
		log.Criticalf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "OK")
}
