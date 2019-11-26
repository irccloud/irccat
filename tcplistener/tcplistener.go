package tcplistener

import (
	"bufio"
	"github.com/irccloud/irccat/dispatcher"
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
	listener.socket, err = net.Listen("tcp", viper.GetString("tcp.listen"))
	if err != nil {
		return nil, err
	}

	return &listener, nil
}

func (l *TCPListener) Run(irccon *irc.Connection) {
	log.Infof("Listening for TCP requests on %s", viper.GetString("tcp.listen"))
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
			dispatcher.Send(l.irc, msg, log, conn.RemoteAddr().String())
		}
	}
	conn.Close()
}
