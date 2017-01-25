package httplistener

import (
	"bytes"
	"github.com/irccloud/irccat/dispatcher"
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

	body := new(bytes.Buffer)
	body.ReadFrom(request.Body)
	message := body.String()

	if message == "" {
		log.Warningf("%s - No message body in POST request", request.RemoteAddr)
		return
	}
	dispatcher.Send(hl.irc, message, log, request.RemoteAddr)
}
