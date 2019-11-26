package main

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"strings"
	"time"
)

func (i *IRCCat) connectIRC(debug bool) error {
	irccon := irc.IRC(viper.GetString("irc.nick"), viper.GetString("irc.realname"))
	i.irc = irccon

	irccon.Debug = debug
	irccon.Timeout = time.Second * 15
	irccon.RequestCaps = []string{"away-notify", "account-notify", "draft/message-tags-0.2"}
	irccon.UseTLS = viper.GetBool("irc.tls")

	if viper.IsSet("irc.sasl_pass") && viper.GetString("irc.sasl_pass") != "" {
		if viper.IsSet("irc.sasl_login") && viper.GetString("irc.sasl_login") != "" {
			irccon.SASLLogin = viper.GetString("irc.sasl_login")
		} else {
			irccon.SASLLogin = viper.GetString("irc.nick")
		}
		irccon.SASLPassword = viper.GetString("irc.sasl_pass")
		irccon.UseSASL = true
	}

	if viper.GetBool("irc.tls_skip_verify") {
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	irccon.Password = viper.GetString("irc.server_pass")

	err := irccon.Connect(viper.GetString("irc.server"))
	if err != nil {
		return err
	}

	irccon.AddCallback("001", i.handleWelcome)
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		msg := event.Message()
		if (msg[0] == '?' || msg[0] == '!') && len(msg) > 1 {
			go i.handleCommand(event)
		}
	})

	irccon.AddCallback("353", i.handleNames)
	irccon.AddCallback("JOIN", i.handleJoin)
	irccon.AddCallback("PART", i.handlePart)
	irccon.AddCallback("QUIT", i.handleQuit)
	irccon.AddCallback("KILL", i.handleQuit)
	irccon.AddCallback("NICK", i.handleNick)

	return nil
}

func (i *IRCCat) handleWelcome(e *irc.Event) {
	log.Infof("Negotiated IRCv3 capabilities: %v", i.irc.AcknowledgedCaps)
	if viper.IsSet("irc.identify_pass") && viper.GetString("irc.identify_pass") != "" {
		i.irc.SendRawf("NICKSERV IDENTIFY %s", viper.GetString("irc.identify_pass"))
	}

	log.Infof("Connected, joining channels...")
	for _, channel := range viper.GetStringSlice("irc.channels") {
		key_var := fmt.Sprintf("irc.keys.%s", channel)
		if strings.ContainsAny(channel, " \t") {
			log.Errorf("Channel name '%s' contains whitespace. Set a channel key by setting the config variable irc.keys.#channel",
				channel)
			continue
		}

		if viper.IsSet(key_var) {
			i.irc.Join(channel + " " + viper.GetString(key_var))
		} else {
			i.irc.Join(channel)
		}
		i.channels.Add(channel)
	}
}
