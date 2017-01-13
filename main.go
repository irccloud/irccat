package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/irccloud/irccat/httplistener"
	"github.com/irccloud/irccat/tcplistener"
	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

var log = loggo.GetLogger("main")

type IRCCat struct {
	control_chan string
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

	irccat := IRCCat{}
	irccat.signals = make(chan os.Signal, 1)
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

	httplistener.New(irccat.irc)

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
	irccon.Debug = true
	irccon.UseTLS = viper.GetBool("irc.tls")
	if viper.GetBool("irc.tls_skip_verify") {
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true, MaxVersion: tls.VersionTLS11}
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

	i.irc = irccon
	return nil
}

func (i *IRCCat) authorisedUser(nick string) bool {
	return false
}

func (i *IRCCat) handleWelcome(e *irc.Event) {
	for _, channel := range viper.GetStringSlice("irc.channels") {
		i.irc.Join(channel)
	}
}

func (i *IRCCat) handleCommand(event *irc.Event) {
	msg := event.Message()
	respond_to := event.Arguments[0]

	if respond_to[0] != '#' && !i.authorisedUser(event.Nick) {
		// Command not in a channel, or not from an authorised user
		log.Infof("Unauthorised command: %s (%s) %s", event.Nick, respond_to, msg)
		return
	}
	log.Infof("Authorised command: %s (%s) %s", event.Nick, respond_to, msg)

	channel := ""
	if respond_to[0] == '#' {
		channel = respond_to
	}

	parts := strings.SplitN(msg, " ", 1)

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(viper.GetString("commands.handler"), event.Nick, channel, respond_to, parts[0][1:])
	} else {
		cmd = exec.Command(viper.GetString("commands.handler"), event.Nick, channel, respond_to, parts[0][1:], parts[1])
	}
	i.runCommand(cmd, respond_to)
}

// Run a command with the output going to the nick/channel identified by respond_to
func (i *IRCCat) runCommand(cmd *exec.Cmd, respond_to string) {
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Running command %s failed: %s", cmd.Args, err)
		i.irc.Privmsgf(respond_to, "Command failed: %s", err)
	}

	lines := strings.Split(out.String(), "\n")
	line_count := len(lines)
	if line_count > viper.GetInt("commands.max_response_lines") {
		line_count = viper.GetInt("commands.max_response_lines")
	}

	for _, line := range lines[0:line_count] {
		if line != "" {
			i.irc.Privmsg(respond_to, line)
		}
	}
}
