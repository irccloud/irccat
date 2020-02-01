# irccat
[![Build Status](https://travis-ci.org/irccloud/irccat.svg?branch=master)](https://travis-ci.org/irccloud/irccat)<a href="https://www.irccloud.com/invite?channel=%23irccat&amp;hostname=irc.irccloud.com&amp;port=6697&amp;ssl=1" target="_blank"><img src="https://img.shields.io/badge/IRC-%23irccat-1e72ff.svg?style=flat"  height="20"></a>

A reimplementation of [irccat](https://github.com/RJ/irccat), the
original ChatOps tool, in Go.

irccat lets you easily send events to IRC channels from scripts and
other applications.

## Installation

Download the [latest
release](https://github.com/irccloud/irccat/releases) from Github, put
the [example config](examples/irccat.json)
in `/etc/irccat.json` or the local directory and customise it, and run!

A Docker container is also [provided on Docker Hub](https://hub.docker.com/r/irccloud/irccat).

## TCP → IRC

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

IRC formatting is supported (see a full [list of codes](dispatcher/colours.go#L5)):

    echo "Status is%GREEN OK %NORMAL" | nc irccat-host 12345

## HTTP → IRC

HTTP listeners are configured by setting keys under `http.listeners`.

### Generic HTTP Endpoint
```json
"generic": true
```

An endpoint for sending messages similar to the TCP port. You can use curl in lieu
of netcat, with `-d @-` to read POST data from stdin, like so:

    echo "Hello world" | curl -d @- http://irccat-host/send

### Generic HTTP Endpoint with authentication

```json
"generic": {
    "secret": "my_secret"
}
```

Adding an optional secret allows you to require a single secret token before sending
messages to the specified channels. (Using HTTPS is recommended to ensure key security)

    echo "Hello world" | curl -H "Authorization: Bearer my_secret" -d @- http://irccat-host/send

### Grafana Webhook
```json
"grafana": "#channel"
```

Grafana alerts can be sent to `/grafana`. They will be sent to the
channel defined in `http.listeners.grafana`. Note that this endpoint is currently
unauthenticated.

### GitHub Webhook
```json
"github": {
	"secret": "my_secret",
	"default_channel": "#channel",
	"repositories": {
	    "irccat": "#irccat-dev"
	}
}
```

Receives GitHub webhooks at `/github`. Currently supports issues, issue comments,
pull requests, pushes, and releases. The webhook needs to be configured to post data
as JSON, not as form-encoded.

The destination channel for notifications from each repository is set in
`http.listeners.github.repositories.repo_name`, where `repo_name` is the name of the
repository, lowercased.

If `http.listeners.github.default_channel` is set, received notifications will be
sent to this channel unless overridden in `http.listeners.github.repositories`. Otherwise,
unrecognised repositories will be ignored.

GitHub can be configured to deliver webhooks to irccat on an organisation level which, combined
with the `default_channel` setting, significantly reduces configuration effort compared to
GitHub's old integrations system.

### Prometheus Alertmanager Webhook
```json
"prometheus": "#channel"
```

Receives [Prometheus Alertmanager webhooks](https://prometheus.io/docs/alerting/configuration/#webhook_config) at `/prometheus`. They will be sent to the channel defined in `http.listeners.prometheus`. Note that this endpoint is unauthenticated.

## IRC → Shell
You can use irccat to execute commands from IRC:

    ?commandname string of arguments

This will call your `commands.handler` script, with the following
environment variables:

* `IRCCAT_COMMAND`: The name of the command, without the preceding `?`
  ("commandname" in this example)
* `IRCCAT_ARGS`: The arguments provided ("string of arguments" in this
  example)
* `IRCCAT_NICK`: Nickname of the calling user
* `IRCCAT_USER`: Username of the calling user
* `IRCCAT_HOST`: Hostname of the calling user
* `IRCCAT_CHANNEL`: Channel the command was issued in (may be blank if
  issued in PM)
* `IRCCAT_RESPOND_TO`: The nick or channel that the STDOUT of the
  command will be sent to
* `IRCCAT_RAW`: The raw IRC line received

The command handler's STDOUT will be sent back to the nick or channel
where the command was issued.

An example python command handler, which dispatches commands to
individual shell scripts, can be found in
[examples/command_handler.py](examples/command_handler.py).

irccat will only recognise commands from users in private message if
the user is joined to `commands.auth_channel` defined in the config.

## Full list of differences from RJ/irccat
* Supports TLS connections to IRC servers.
* HTTP endpoint handlers.
* Doesn't support !join, !part commands, but does automatically reload
  the config and join new channels.
* Arguments are passed as environment variables to the command handler
  script, rather than as a single argument.
