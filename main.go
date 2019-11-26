package main

import (
	"flag"
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/fsnotify/fsnotify"
	"github.com/irccloud/irccat/httplistener"
	"github.com/irccloud/irccat/tcplistener"
	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"os"
	"os/signal"
	"syscall"
)

var log = loggo.GetLogger("main")

var branch string
var revision string

type IRCCat struct {
	auth_channel string
	channels     mapset.Set
	auth_users   map[string]bool
	irc          *irc.Connection
	tcp          *tcplistener.TCPListener
	signals      chan os.Signal
}

func main() {
	debug := flag.Bool("debug", false, "Print raw IRC lines")
	flag.Parse()

	loggo.ConfigureLoggers("<root>=INFO")
	log.Infof("IRCCat %s (%s) starting...", branch, revision)
	viper.SetConfigName("irccat")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	var err error

	err = viper.ReadInConfig()
	if err != nil {
		log.Errorf("Error reading config file - exiting. I'm looking for irccat.[json|yaml|toml|hcl] in . or /etc")
		return
	}

	irccat := IRCCat{auth_users: map[string]bool{},
		signals:      make(chan os.Signal, 1),
		channels:     mapset.NewSet(),
		auth_channel: viper.GetString("commands.auth_channel")}

	viper.WatchConfig()
	viper.OnConfigChange(irccat.handleConfigChange)

	signal.Notify(irccat.signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go irccat.signalHandler()

	irccat.tcp, err = tcplistener.New()
	if err != nil {
		log.Criticalf("Error starting TCP listener: %s", err)
		return
	}

	err = irccat.connectIRC(*debug)

	if err != nil {
		log.Criticalf("Error connecting to IRC server: %s", err)
		return
	}

	if viper.IsSet("http") {
		httplistener.New(irccat.irc)
	}

	irccat.tcp.Run(irccat.irc)
	irccat.irc.Loop()
}

func (i *IRCCat) signalHandler() {
	sig := <-i.signals
	log.Infof("Exiting on %s", sig)
	i.irc.QuitMessage = fmt.Sprintf("Exiting on %s", sig)
	i.irc.Quit()
}

func (i *IRCCat) handleConfigChange(e fsnotify.Event) {
	log.Infof("Reloaded config")

	new_channels := mapset.NewSet()

	for _, channel := range viper.GetStringSlice("irc.channels") {
		new_channels.Add(channel)
		if !i.channels.Contains(channel) {
			log.Infof("Joining new channel %s", channel)
			i.irc.Join(channel)
			i.channels.Add(channel)
		}
	}

	it := i.channels.Difference(new_channels).Iterator()
	for channel := range it.C {
		log.Infof("Leaving channel %s", channel)
		i.irc.Part(channel.(string))
		i.channels.Remove(channel)
	}
}
