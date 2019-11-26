package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
)

func (i *IRCCat) handleCommand(event *irc.Event) {
	msg := event.Message()
	channel := ""
	respond_to := event.Arguments[0]

	if i.inChannel(respond_to) {
		channel = respond_to
	} else {
		respond_to = event.Nick
		if !i.authorisedUser(event.Nick) {
			log.Infof("Unauthorised command: %s (%s) %s", event.Nick, respond_to, msg)
			return
		}
	}

	log.Infof("Authorised command: %s (%s) %s", event.Nick, respond_to, msg)

	parts := strings.SplitN(msg, " ", 2)

	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	handler := viper.GetString("commands.handler")
	if handler != "" {
		cmd := exec.Command(handler)
		cmd.Env = append(os.Environ(), fmt.Sprintf("IRCCAT_NICK=%s", event.Nick),
			fmt.Sprintf("IRCCAT_USER=%s", event.User),
			fmt.Sprintf("IRCCAT_HOST=%s", event.Host),
			fmt.Sprintf("IRCCAT_CHANNEL=%s", channel),
			fmt.Sprintf("IRCCAT_RESPOND_TO=%s", respond_to),
			fmt.Sprintf("IRCCAT_COMMAND=%s", parts[0][1:]),
			fmt.Sprintf("IRCCAT_ARGS=%s", args),
			fmt.Sprintf("IRCCAT_RAW=%s", event.Raw))

		i.runCommand(cmd, respond_to)
	}
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
			// 360 bytes is the worst-case maximum size for PRIVMSG lines. Truncate the lines at that length.
			if len(line) > 360 {
				line = line[:360]
			}
			i.irc.Privmsg(respond_to, line)
		}
		// Pause between lines to avoid being flooded out
		time.Sleep(250)
	}
}
