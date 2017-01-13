package main

import (
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"strings"
)

func (i *IRCCat) authorisedUser(nick string) bool {
	_, exists := i.auth_users[nick]
	return exists
}

func (i *IRCCat) handleJoin(e *irc.Event) {
	if e.Arguments[0] == viper.GetString("commands.auth_channel") {
		i.auth_users[e.Nick] = true
	}
}

func (i *IRCCat) handlePart(e *irc.Event) {
	if e.Arguments[0] == viper.GetString("commands.auth_channel") {
		delete(i.auth_users, e.Nick)
	}
}

func (i *IRCCat) handleNames(e *irc.Event) {
	if e.Arguments[2] == viper.GetString("commands.auth_channel") {
		nicks := strings.Split(e.Arguments[3], " ")
		for _, nick := range nicks {
			i.auth_users[nick] = true
		}
	}
}
