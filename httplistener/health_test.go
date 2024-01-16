package httplistener

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"testing"
	"time"

	"net/http"

	"github.com/irccloud/irccat/util"
	"github.com/juju/loggo"
	"github.com/spf13/viper"
)

var origLag = LagInterval

var configFmt = `---
irc:
  health_file: %s
  health_period: 1ms
http:
  health_endpoint: /testing/healthz
`

func startServer(t *testing.T) {
	hl := HTTPListener{
		http: http.Server{Addr: "localhost:18045"},
	}
	http.HandleFunc(viper.GetString("http.health_endpoint"), hl.healthHandler)
	go hl.http.ListenAndServe()
	t.Cleanup(func() {hl.http.Shutdown(context.Background())})
}

func getOne(t *testing.T) (*http.Response, string) {
	res, err := http.Get("http://localhost:18045/testing/healthz")
	if err != nil {
		t.Error(err)
	}
	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	err = res.Body.Close()
	if err != nil {
		t.Error(err)
	}
	return res, string(got)
}

func TestHealthHandler(t *testing.T) {
	writer, err := loggo.RemoveWriter("default")
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() {loggo.DefaultContext().AddWriter("default", writer)})
	LagInterval = 0
	t.Cleanup(func() {LagInterval = origLag})
	dir := t.TempDir()
	now := time.Now()
	file := path.Join(dir, "timestamp")
	if err := util.WriteTimestamp(file, now); err != nil {
		t.Error(err)
	}
	viper.SetConfigType("yaml")
	config := []byte(fmt.Sprintf(configFmt, file))
	viper.ReadConfig(bytes.NewBuffer(config))
	startServer(t)
	time.Sleep(time.Millisecond)
	// Fail
	resp, got := getOne(t)
	if resp.StatusCode != 500 {
		t.Error("unexpected status", resp.Status)
	}
	t.Log(resp.Status, got)
	// Success
	viper.Set("irc.health_period", time.Second)
	resp, got = getOne(t)
	if resp.StatusCode != 200 {
		t.Error("unexpected failure", resp.Status)
	}
	if string(got) != "OK\n" {
		t.Error("unexpected output", string(got))
	}
}
