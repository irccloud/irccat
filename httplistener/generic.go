package httplistener

import (
	"bytes"
	"fmt"
	"github.com/irccloud/irccat/dispatcher"
	"github.com/spf13/viper"
	"net/http"
)

// Examples of using curl to post to /send.
//
// echo "Hello, world" | curl -d @- http://irccat.example.com/send
// echo "#test,@alice Hello, world" | curl -d @- http://irccat.example.com/send
//
func (hl *HTTPListener) genericHandler(w http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.NotFound(w, request)
		return
	}

	// Optional simple auth via token
	secret := viper.GetString("http.listeners.generic.secret")
	if secret != "" {
		auth := request.Header.Get("Authorization")
		expecting := fmt.Sprintf("Bearer %s", secret)
		if auth != expecting {
			http.Error(w, "Invalid Authorization", http.StatusUnauthorized)
			log.Warningf("%s - Invalid Authorization!", request.RemoteAddr)
			return
		}
	}

	body := new(bytes.Buffer)
	body.ReadFrom(request.Body)
	message := body.String()

	if message == "" {
		log.Warningf("%s - No message body in POST request", request.RemoteAddr)
		return
	}
	dispatcher.Send(hl.irc, message, log, request.RemoteAddr)
}
