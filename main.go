package main

import (
	"crypto/tls"
	"fmt"
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

type IRCCat struct {
	control_chan string
	auth_users   map[string]bool
	irc          *irc.Connection
	tcp          *tcplistener.TCPListener
	signals      chan os.Signal
}

func main() {
	loggo.ConfigureLoggers("<root>=DEBUG")
	log.Infof("IRCCat starting...")
	viper.SetConfigName("irccat")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	var err error

	err = viper.ReadInConfig()
	if err != nil {
		log.Errorf("Error reading config file - exiting. I'm looking for irccat.[json|yaml|toml|hcl] in . or /etc")
		return
	}
	irccat := IRCCat{auth_users: map[string]bool{}, signals: make(chan os.Signal, 1)}

	signal.Notify(irccat.signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go irccat.signalHandler()

	irccat.tcp, err = tcplistener.New()
	if err != nil {
		return
	}

	err = irccat.connectIRC()

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

func (i *IRCCat) connectIRC() error {
	irccon := irc.IRC(viper.GetString("irc.nick"), viper.GetString("irc.realname"))
	irccon.UseTLS = viper.GetBool("irc.tls")
	if viper.GetBool("irc.tls_skip_verify") {
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	err := irccon.Connect(viper.GetString("irc.server"))
	if err != nil {
		return err
	}

	irccon.AddCallback("001", i.handleWelcome)
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		if event.Message()[0] == '?' || event.Message()[0] == '!' {
			go i.handleCommand(event)
		}
	})

	irccon.AddCallback("353", i.handleNames)
	irccon.AddCallback("JOIN", i.handleJoin)
	irccon.AddCallback("PART", i.handlePart)
	irccon.AddCallback("QUIT", i.handlePart)

	i.irc = irccon
	return nil
}

func (i *IRCCat) handleWelcome(e *irc.Event) {
	for _, channel := range viper.GetStringSlice("irc.channels") {
		i.irc.Join(channel)
	}
}
