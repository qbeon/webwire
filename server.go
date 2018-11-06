package webwire

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/pkg/errors"
	"github.com/qbeon/webwire-go/message"
	"github.com/valyala/fasthttp"
)

// server represents a headless WebWire server instance,
// where headless means there's no HTTP server that's hosting it
type server struct {
	impl              ServerImplementation
	httpServer        *fasthttp.Server
	listener          net.Listener
	tlsConfig         *tls.Config
	sessionManager    SessionManager
	sessionKeyGen     SessionKeyGenerator
	sessionInfoParser SessionInfoParser

	// State
	addr            url.URL
	options         ServerOptions
	configMsg       []byte
	certFilePath    string
	keyFilePath     string
	shutdown        bool
	shutdownRdy     chan bool
	currentOps      uint32
	opsLock         *sync.Mutex
	connectionsLock *sync.Mutex
	connections     []*connection
	sessionsEnabled bool
	sessionRegistry *sessionRegistry
	messagePool     message.Pool

	// Internals
	upgrader websocket.FastHTTPUpgrader
	warnLog  *log.Logger
	errorLog *log.Logger
}

func (srv *server) shutdownHTTPServer() error {
	if srv.httpServer == nil {
		return nil
	}
	if err := srv.httpServer.Shutdown(); err != nil {
		return fmt.Errorf("Couldn't properly shutdown HTTP server: %s", err)
	}
	return nil
}

// Run implements the Server interface
func (srv *server) Run() error {
	if srv.tlsConfig != nil {
		if err := srv.httpServer.ServeTLS(
			tcpKeepAliveListener{srv.listener.(*net.TCPListener)},
			srv.certFilePath,
			srv.keyFilePath,
		); err != io.EOF {
			return errors.Wrap(err, "HTTPS server failure")
		}
	} else {
		if err := srv.httpServer.Serve(
			tcpKeepAliveListener{srv.listener.(*net.TCPListener)},
		); err != io.EOF {
			return errors.Wrap(err, "HTTP server failure")
		}
	}

	return nil
}

// Address implements the Server interface
func (srv *server) Address() string {
	return srv.addr.String()
}

// AddressURL implements the Server interface
func (srv *server) AddressURL() url.URL {
	return srv.addr
}

// Shutdown implements the Server interface
func (srv *server) Shutdown() error {
	srv.opsLock.Lock()
	srv.shutdown = true
	// Don't block if there's no currently processed operations
	if srv.currentOps < 1 {
		srv.opsLock.Unlock()
		return srv.shutdownHTTPServer()
	}
	srv.opsLock.Unlock()
	<-srv.shutdownRdy

	return srv.shutdownHTTPServer()
}

// ActiveSessionsNum implements the Server interface
func (srv *server) ActiveSessionsNum() int {
	return srv.sessionRegistry.activeSessionsNum()
}

// SessionConnectionsNum implements the Server interface
func (srv *server) SessionConnectionsNum(sessionKey string) int {
	return srv.sessionRegistry.sessionConnectionsNum(sessionKey)
}

// SessionConnections implements the Server interface
func (srv *server) SessionConnections(sessionKey string) []Connection {
	connections := srv.sessionRegistry.sessionConnections(sessionKey)
	if connections == nil {
		return nil
	}
	list := make([]Connection, len(connections))
	i := 0
	for connection := range connections {
		list[i] = connection
		i++
	}
	return list
}

// CloseSession implements the Server interface
func (srv *server) CloseSession(sessionKey string) (
	affectedConnections []Connection,
	errors []error,
	generalError error,
) {
	connections := srv.sessionRegistry.sessionConnections(sessionKey)

	errors = make([]error, len(connections))
	if connections == nil {
		return nil, nil, nil
	}
	affectedConnections = make([]Connection, len(connections))
	i := 0
	errNum := 0
	for connection := range connections {
		affectedConnections[i] = connection
		err := connection.CloseSession()
		if err != nil {
			errors[i] = err
			errNum++
		} else {
			errors[i] = nil
		}
		i++
	}

	if errNum > 0 {
		generalError = fmt.Errorf(
			"%d errors during the closure of a session",
			errNum,
		)
	}

	return affectedConnections, errors, generalError
}
