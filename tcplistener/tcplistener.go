package tcplistener

import (
	"bufio"
	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
	"net"
	"strings"
)

var log = loggo.GetLogger("TCPListener")

type TCPListener struct {
	socket net.Listener
	irc    *irc.Connection
}

func New() (*TCPListener, error) {
	var err error

	listener := TCPListener{}
	listener.socket, err = net.Listen("tcp", viper.GetString("tcp_listen"))
	if err != nil {
		return nil, err
	}

	return &listener, nil
}

func (l *TCPListener) Run(irccon *irc.Connection) {
	log.Infof("Listening for TCP requests on %s", viper.GetString("tcp_listen"))
	l.irc = irccon
	go l.acceptConnections()
}

func (l *TCPListener) acceptConnections() {
	for {
		conn, err := l.socket.Accept()
		if err != nil {
			break
		}
		go l.handleConnection(conn)
	}
	l.socket.Close()
}

func (l *TCPListener) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		msg = strings.Trim(msg, "\r\n")
		if len(msg) > 0 {
			log.Infof("[%s] message: %s", conn.RemoteAddr(), msg)
			l.parseMessage(msg)
		}
	}
}

func (l *TCPListener) parseMessage(msg string) {
	channels := viper.GetStringSlice("irc.channels")

	if msg[0] == '#' || msg[0] == '@' {
		parts := strings.SplitN(msg, " ", 2)
		if parts[0] == "#*" {
			for _, channel := range channels {
				chan_parts := strings.Split(channel, " ")
				l.irc.Privmsg(chan_parts[0], replaceFormatting(parts[1]))
			}
		} else {
			targets := strings.Split(parts[0], ",")
			for _, target := range targets {
				if target[0] == '@' {
					target = target[1:]
				}
				l.irc.Privmsg(target, replaceFormatting(parts[1]))
			}
		}
	} else if len(msg) > 7 && msg[0:6] == "%TOPIC" {
		parts := strings.SplitN(msg, " ", 3)
		l.irc.SendRawf("TOPIC %s :%s", parts[1], replaceFormatting(parts[2]))
	} else {
		if len(channels) > 0 {
			chan_parts := strings.Split(channels[0], " ")
			l.irc.Privmsg(chan_parts[0], replaceFormatting(msg))
		}
	}
}
