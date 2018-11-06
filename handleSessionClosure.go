package webwire

import (
	"github.com/qbeon/webwire-go/message"
	"github.com/qbeon/webwire-go/wwrerr"
)

// handleSessionClosure handles session destruction requests
// and returns an error if the ongoing connection cannot be proceeded
func (srv *server) handleSessionClosure(
	conn *connection,
	msg *message.Message,
) {
	if !srv.sessionsEnabled {
		srv.failMsg(conn, msg, wwrerr.SessionsDisabledErr{})
		return
	}

	if !conn.HasSession() {
		// Send confirmation even though no session was closed
		srv.fulfillMsg(conn, msg, Payload{})
		return
	}

	// Deregister session from active sessions registry
	srv.sessionRegistry.deregister(conn)

	// Synchronize session destruction to the client
	if err := conn.notifySessionClosed(); err != nil {
		srv.failMsg(conn, msg, nil)
		srv.errorLog.Printf("CRITICAL: Internal server error, "+
			"couldn't notify client about the session destruction: %s",
			err,
		)
		return
	}

	// Reset the session on the connection
	conn.setSession(nil)

	// Send confirmation
	srv.fulfillMsg(conn, msg, Payload{})
}
