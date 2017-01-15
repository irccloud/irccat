# irccat
[![Build Status](https://travis-ci.org/irccloud/irccat.svg?branch=master)](https://travis-ci.org/irccloud/irccat)

A reimplementation of [irccat](https://github.com/RJ/irccat), the
original ChatOps tool, in Go. irccat lets you easily send events
to IRC channels from scripts and other applications.

## TCP → IRC
Just cat a string to the TCP port - it'll be sent to the first channel
defined in your channel list:

    echo "Hello world" | nc -q 0 irccat-host 12345

Or specify a channel or nickname to send to:

    echo "#channel Hello world" | nc -q 0 irccat-host 12345
    echo "@nick Hello world" | nc -q 0 irccat-host 12345

IRC formatting is supported (see a full [list of
codes](https://github.com/irccloud/irccat/blob/master/tcplistener/colours.go#L5)):

    echo "Status is%GREEN OK %NORMAL" | nc -q 0 irccat-host 12345a

## HTTP → IRC
There's a simple HTTP endpoint for sending messages:

    curl -X POST http://irccat-host:8045/send -d
        '{"to": "#channel", "body": "Hello world"}

There are also endpoints which support app-specific webhooks, currently:

* Grafana alerts can be sent to `/grafana`. They will be sent to the
  channel defined in `http.listeners.grafana`.

Note that there is (currently) no authentication on the HTTP endpoints,
so you should make sure you firewall them from the world.

## IRC → Shell
You can use irccat to execute commands from IRC:

    ?commandname arguments

This will call your `commands.handler` script with the command-line
arguments:

    nickname, [channel], respond_to, commandname, [arguments]

irccat will only recognise commands from users in private message if
the user is joined to `commands.auth_channel` defined in the config.

## Full list of differences from RJ/irccat
* Supports TLS connections to IRC servers.
* HTTP endpoint handlers.
* Doesn't support !join, !part commands, but does automatically reload
  the config and join new channels.
* Arguments are passed individually to the command handler script,
  rather than in one string as a single argument.
