package dispatcher

import (
	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"strings"
)

// Take a string, parse out the recipients, and send to IRC.
//
// eg:
//  hello world                 [goes to default channel]
//  #test hello world           [goes to #test, if joined]
//  #test,@alice hello world    [goes to #test and alice]
//  #* hello world              [goes to all channels bot is in]
func Send(irc *irc.Connection, msg string, log loggo.Logger, origin string) {
	channels := viper.GetStringSlice("irc.channels")

	if msg[0] == '#' || msg[0] == '@' {
		parts := strings.SplitN(msg, " ", 2)
		if parts[0] == "#*" {
			for _, channel := range channels {
				irc.Privmsg(channel, replaceFormatting(parts[1]))
			}
		} else {
			targets := strings.Split(parts[0], ",")
			for _, target := range targets {
				if target[0] == '@' {
					target = target[1:]
				}
				irc.Privmsg(target, replaceFormatting(parts[1]))
			}
		}
		log.Infof("from[%s] send[%s] %s", origin, parts[0], parts[1])
	} else if len(msg) > 7 && msg[0:6] == "%TOPIC" {
		parts := strings.SplitN(msg, " ", 3)
		irc.SendRawf("TOPIC %s :%s", parts[1], replaceFormatting(parts[2]))
		log.Infof("from[%s] topic[%s] %s", origin, parts[1], parts[2])
	} else {
		if len(channels) > 0 {
			irc.Privmsg(channels[0], replaceFormatting(msg))
			log.Infof("from[%s] send_default[%s] %s", origin, channels[0], msg)
		}
	}
}
