// Some clients refuse to send payloads that don't match the Content-Type. In
// other words, trying to send "%BOLD hw" or "\x02 hw", as written, wouldn't
// be allowed for Content-Type "application/x-www-form-urlencoded" without
// first being encoded. But irccat by default doesn't do any decoding. Thus,
// when faced with one of these problematic clients, specify config option
// http.listeners.generic.strict to get the behavior shown here:
//
// mismatch
//
//  $ echo "%BOLDhw" | curl -d @- http://localhost/send
//  400 Bad Request
//
// urlencoded
//
//  $ echo "%BOLDhw" | curl --data-urlencode  @- http://localhost/send
//  200 OK
//
// urlencoded non-printable
//
//  $ printf "\x02hw" | curl --data-urlencode  @- http://localhost/send
//  200 OK
//
// octetstream
//
//  $ echo "%BOLDhw" | curl --data-binary @- \
//    -H 'Content-Type: application/octet-stream' http://localhost/send
//  200 OK
//
// multipart quoted-printable
//
//  $ echo '%BOLDhw' | curl -F 'foo=@-;encoder=quoted-printable' \
//    http://localhost/send
//  200 OK
//
// multipart 8bit
//
//  $ echo '%BOLDhw' | curl -F 'foo=@-;encoder=8bit' http://localhost/send
//  200 OK
//
// multipart base64
//
//  $ echo '%BOLDhw' | curl -F 'foo=@-;encoder=base64' http://localhost/send
//  200 OK
//
// The gist is that when strict mode is active, popular encodings will work
// while mismatches won't, even though they may still appear to at times.
//
package httplistener

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/juju/loggo"
	"github.com/spf13/viper"
	irc "github.com/thoj/go-ircevent"
)

var genericTestListen = "localhost:18045"

func genericTestStartHTTPServer(t *testing.T, endpoint string) {
	hl := HTTPListener{
		http: http.Server{Addr: genericTestListen},
	}

	http.HandleFunc(endpoint, hl.genericHandler)
	go hl.http.ListenAndServe()
	t.Cleanup(func() {hl.http.Shutdown(context.Background());})
	time.Sleep(time.Millisecond)
}

func genericTestSendOutput(message []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", genericTestListen)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(message)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 1024)
	_ , err = io.ReadAtLeast(conn, b, 24)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func runGeneric(t *testing.T, reqFileName string) (string, string) {
	var message string
	origSender := genericSender
	genericSender = func(_ *irc.Connection, m string, _ loggo.Logger, _ string) {
		message = m
	}
	t.Cleanup(func(){genericSender = origSender})

	src, err := os.ReadFile(path.Join("testdata", reqFileName))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := genericTestSendOutput(src)
	if err != nil {
		t.Fatal(err)
	}
	return message, string(resp)
}

// Non-strict

func testGenericBaseline(t *testing.T) {
	message, resp := runGeneric(t, "mismatch")
	if message != "%BOLDhw" {
		t.Fatalf("Expected %q, got: %q", "%BOLDhw", message)
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 200 OK") {
		t.Fatalf("Unexpected message: %s", resp)
	}
}

// Strict

func testGenericStrict(t *testing.T) {
	message, resp := runGeneric(t, "mismatch")
	if message != "" {
		t.Fatalf("Expected %q, got: %q", "", message)
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 400 Bad Request") {
		t.Fatalf("Unexpected message: %s", resp)
	}
}

func testURLEncoded(t *testing.T) {
	message, resp := runGeneric(t, "urlencoded")
	if message != "%BOLDhw\n" { // Note the linefeed
		t.Fatalf("Expected %q, got: %q", "%BOLDhw\n", message)
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 200 OK") {
		t.Fatalf("Unexpected message: %s", resp)
	}
}

func testURLEncodedNonPrintable(t *testing.T) {
	message, resp := runGeneric(t, "urlencoded_npc")
	if message != "\x02hw" {
		t.Fatalf("Expected %q, got: %q", "\x02hw", message)
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 200 OK") {
		t.Fatalf("Unexpected message: %s", resp)
	}
}

func testOctetStream(t *testing.T) {
	message, resp := runGeneric(t, "octetstream")
	if message != "%BOLDhw\n" {
		t.Fatalf("Expected %q, got: %q", "%BOLDhw\n", message)
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 200 OK") {
		t.Fatalf("Unexpected message: %s", resp)
	}
}

func testMultipartQP(t *testing.T) {
	message, resp := runGeneric(t, "multipart_qp")
	if message != "%BOLDhw\n" {
		t.Fatalf("Expected %q, got: %q", "%BOLDhw\n", message)
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 200 OK") {
		t.Fatalf("Unexpected message: %s", resp)
	}
}

func testMultipart8bit(t *testing.T) {
	message, resp := runGeneric(t, "multipart_8bit")
	if message != "%BOLDhw\n" {
		t.Fatalf("Expected %q, got: %q", "%BOLDhw\n", message)
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 200 OK") {
		t.Fatalf("Unexpected message: %s", resp)
	}
}

func testMultipartBase64(t *testing.T) {
	message, resp := runGeneric(t, "multipart_base64")
	if message != "%BOLDhw\n" {
		t.Fatalf("Expected %q, got: %q", "%BOLDhw\n", message)
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 200 OK") {
		t.Fatalf("Unexpected message: %s", resp)
	}
}

func TestGeneric(t *testing.T) {
	writer, err := loggo.RemoveWriter("default")
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() {loggo.DefaultContext().AddWriter("default", writer)})
	genericTestStartHTTPServer(t, "/send")

	t.Run("Baseline", testGenericBaseline)

	// Turn on strict for the rest of these
	viper.Set("http.listeners.generic.strict", true)

	t.Run("Strict Mismatch", testGenericStrict)
	t.Run("Strict URL Encoded", testURLEncoded)
	t.Run("Strict URL Encoded Non-printable", testURLEncodedNonPrintable)
	t.Run("Strict Octet Stream", testOctetStream)
	t.Run("Strict Multipart Quoted Printable", testMultipartQP)
	t.Run("Strict Multipart 8bit", testMultipart8bit)
	t.Run("Strict Multipart Base 64", testMultipartBase64)

	// Restore to zero value
	viper.Set("http.listeners.generic.strict", false)
}
