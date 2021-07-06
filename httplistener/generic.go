package httplistener

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/irccloud/irccat/dispatcher"
	"github.com/spf13/viper"
)

func handleUrlEncodedPostForm(request *http.Request) string {
	request.ParseForm()
	parts := []string{}
	for key, val := range request.PostForm {
		if key != "" {
			parts = append(parts, key)
		}
		for _, v := range val {
			if v != "" {
				parts = append(parts, v)
			}
		}
	}
	return strings.Join(parts, " ")
}

// handleMixed blindly concatenates all bodies in a mixed message.
//
// Headers are discarded. Binary data, including illegal IRC chars, are
// retained as is. Quoted-printable and base64 encodings are recognized.
func handleMixed(request *http.Request) (string, error) {
	mr, err := request.MultipartReader()
	if err != nil {
		return "", err
	}
	parts := []string{}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		b, err := io.ReadAll(p)
		if err != nil {
			return "", err
		}
		if len(b) != 0 {
			if p.Header.Get("content-transfer-encoding") == "base64" {
				encoder := base64.StdEncoding
				if decoded, err := encoder.DecodeString(string(b)); err != nil {
					return "", err
				} else if len(decoded) > 0 {
					b = decoded
				}
			}
			parts = append(parts, string(b))
		}
	}
	return strings.Join(parts, " "), nil
}

var genericSender = dispatcher.Send

// Examples of using curl to post to /send.
//
// echo "Hello, world" | curl -d @- http://irccat.example.com/send
// echo "#test,@alice Hello, world" | curl -d @- http://irccat.example.com/send
//
// See httplistener/generic_tests.go for info on strict mode, which behaves
// differently and is enabled by config option http.listeners.generic.strict
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
	var message string
	strict := viper.GetBool("http.listeners.generic.strict")
	if strict {
		contentType := request.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			message = handleUrlEncodedPostForm(request)
		} else if strings.HasPrefix(contentType, "multipart/") {
			if msg, err := handleMixed(request); err == nil {
				message = msg // otherwise message is "", which triggers 400
			}
		}
	}

	if message == "" {
		body := new(bytes.Buffer)
		body.ReadFrom(request.Body)
		message = body.String()
	}

	if message == "" {
		if strict {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
		log.Warningf("%s - No message body in POST request", request.RemoteAddr)
		return
	}
	genericSender(hl.irc, message, log, request.RemoteAddr)
}
