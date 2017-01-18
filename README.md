# irccat
[![Build Status](https://travis-ci.org/irccloud/irccat.svg?branch=master)](https://travis-ci.org/irccloud/irccat)

A reimplementation of [irccat](https://github.com/RJ/irccat), the
original ChatOps tool, in Go.

irccat lets you easily send events to IRC channels from scripts and
other applications.

## Installation

Download the [latest
release](https://github.com/irccloud/irccat/releases) from Github, put
the [example
config](https://github.com/irccloud/irccat/blob/master/examples/irccat.json)
in `/etc/irccat.json` or the local directory and customise it, and run!

## TCP → IRC
Just cat a string to the TCP port - it'll be sent to the first channel
defined in your channel list:

    echo "Hello world" | nc irccat-host 12345

Or specify a channel or nickname to send to:

    echo "#channel Hello world" | nc irccat-host 12345
    echo "@nick Hello world" | nc irccat-host 12345

You can also send to multiple recipients at once:

    echo "#channel,@nick Hello world | nc irccat-host 12345

And set a channel topic:

    echo "%TOPIC #channel Channel topic" | nc irccat-host 12345

IRC formatting is supported (see a full [list of
codes](https://github.com/irccloud/irccat/blob/master/tcplistener/colours.go#L5)):

    echo "Status is%GREEN OK %NORMAL" | nc irccat-host 12345

## HTTP → IRC
There's a simple HTTP endpoint for sending messages:

    curl -X POST http://irccat-host:8045/send -d
        '{"to": "#channel", "body": "Hello world"}'

There are also endpoints which support app-specific webhooks, currently:

* Grafana alerts can be sent to `/grafana`. They will be sent to the
  channel defined in `http.listeners.grafana`.

More HTTP listeners welcome!

Note that there is (currently) no authentication on the HTTP endpoints,
so you should make sure you firewall them from the world.

## IRC → Shell
You can use irccat to execute commands from IRC:

    ?commandname string of arguments

This will call your `commands.handler` script, with the following
environment variables:

* `IRCCAT_COMMAND`: The name of the command, without the preceding `?`
* `IRCCAT_ARGS`: The arguments provided ("string of arguments" in this
  example)
* `IRCCAT_NICK`: Nickname of the calling user
* `IRCCAT_USER`: Username of the calling user
* `IRCCAT_HOST`: Hostname of the calling user
* `IRCCAT_CHANNEL`: Channel the command was issued in (may be blank if
  issued in PM)
* `IRCCAT_RESPOND_TO`: The nick or channel that the STDOUT of the
  command will be sent to

The command handler's STDOUT will be sent back to the nick or channel
where the command was issued.

irccat will only recognise commands from users in private message if
the user is joined to `commands.auth_channel` defined in the config.

## Full list of differences from RJ/irccat
* Supports TLS connections to IRC servers.
* HTTP endpoint handlers.
* Doesn't support !join, !part commands, but does automatically reload
  the config and join new channels.
* Arguments are passed as environment variables to the command handler
  script, rather than as a single argument.
