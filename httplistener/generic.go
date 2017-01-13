package httplistener

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type genericMessage struct {
	To   string
	Body string
}

func (hl *HTTPListener) genericHandler(w http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.NotFound(w, request)
		return
	}

	var message genericMessage
	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	json.Unmarshal(buf.Bytes(), &message)

	log.Infof("[%s] %s", message.To, message.Body)

	if message.To != "" && message.Body != "" {
		hl.irc.Privmsgf(message.To, message.Body)
	}
}
