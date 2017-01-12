package main

import (
	"crypto/tls"
	"github.com/irccloud/irccat/httplistener"
	"github.com/irccloud/irccat/tcplistener"
	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
)

var log = loggo.GetLogger("main")

type IRCCat struct {
	irc *irc.Connection
	tcp *tcplistener.TCPListener
}

func main() {
	loggo.ConfigureLoggers("<root>=DEBUG")
	viper.SetConfigName("irccat")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	var err error

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	irccat := IRCCat{}

	irccat.tcp, err = tcplistener.New()
	if err != nil {
		return
	}

	err = irccat.connectIRC()

	if err != nil {
		log.Criticalf("Error connecting to IRC server: %s", err)
		return
	}

	httplistener.New(irccat.irc)

	irccat.tcp.Run(irccat.irc)
	irccat.irc.Loop()
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

	i.irc = irccon
	return nil
}

func (i *IRCCat) handleWelcome(e *irc.Event) {
	for _, channel := range viper.GetStringSlice("irc.channels") {
		i.irc.Join(channel)
	}
}
