package fasthttp

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/qbeon/webwire-go/message"
	"github.com/qbeon/webwire-go/wwrerr"
)

// Socket implements the webwire.Socket interface using
// the fasthttp/websocket library
type Socket struct {
	connected bool
	lock      *sync.Mutex
	readLock  *sync.Mutex
	writeLock *sync.Mutex
	conn      *websocket.Conn
	dialer    websocket.Dialer
}

// NewConnectedSocket creates a new fasthttp/websocket based socket
// instance
func NewConnectedSocket(conn *websocket.Conn) *Socket {
	connected := conn != nil
	return &Socket{
		connected: connected,
		lock:      &sync.Mutex{},
		readLock:  &sync.Mutex{},
		writeLock: &sync.Mutex{},
		conn:      conn,
	}
}

// Dial implements the webwire.Socket interface
func (sock *Socket) Dial(serverAddr url.URL, deadline time.Time) (err error) {
	sock.lock.Lock()
	if sock.connected {
		sock.lock.Unlock()
		return errors.New("already connected")
	}
	connection, _, err := sock.dialer.Dial(serverAddr.String(), nil)
	if err != nil {
		sock.lock.Unlock()
		return wwrerr.DisconnectedErr{
			Cause: fmt.Errorf("dial failure: %s", err),
		}
	}
	sock.conn = connection
	sock.connected = true
	sock.lock.Unlock()
	return nil
}

// fasthttpWriteCloser implements the io.WriteCloser interface
type fasthttpWriteCloser struct {
	writer  io.WriteCloser
	onClose func()
}

// Write implements the io.Writer interface
func (wc *fasthttpWriteCloser) Write(p []byte) (n int, err error) {
	return wc.writer.Write(p)
}

// Close implements the io.Closer interface
func (wc *fasthttpWriteCloser) Close() error {
	err := wc.writer.Close()
	wc.onClose()
	return err
}

// GetWriter implements the webwire.Socket interface
func (sock *Socket) GetWriter() (io.WriteCloser, error) {
	sock.writeLock.Lock()

	// Check connection status
	if !sock.IsConnected() {
		return nil, wwrerr.DisconnectedErr{
			Cause: fmt.Errorf("can't write to a closed socket"),
		}
	}

	writer, err := sock.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		sock.writeLock.Unlock()
		return nil, err
	}
	return &fasthttpWriteCloser{
		writer: writer,
		onClose: func() {
			// Unlock the writer lock on writer closure
			sock.writeLock.Unlock()
		},
	}, nil
}

// Read implements the webwire.Socket interface
func (sock *Socket) Read(
	msg *message.Message,
	deadline time.Time,
) wwrerr.SockReadErr {
	sock.readLock.Lock()

	// Check connection status
	if !sock.IsConnected() {
		return SockReadErr{cause: wwrerr.DisconnectedErr{
			Cause: fmt.Errorf("can't read closed socket"),
		}}
	}

	if err := sock.conn.SetReadDeadline(deadline); err != nil {
		sock.readLock.Unlock()
		return SockReadErr{cause: errors.New("couldn't set read deadline")}
	}
	messageType, reader, err := sock.conn.NextReader()
	if err != nil {
		sock.Close()
		sock.readLock.Unlock()
		return SockReadErr{cause: err}
	}

	// Stop deadline timer
	if err := sock.conn.SetReadDeadline(time.Time{}); err != nil {
		sock.readLock.Unlock()
		return SockReadErr{cause: err}
	}

	// Discard message in case of unexpected message types
	if messageType != websocket.BinaryMessage {
		io.Copy(ioutil.Discard, reader)
		sock.readLock.Unlock()
		return SockReadWrongMsgTypeErr{messageType: messageType}
	}

	// Try to read the socket into the buffer
	typeParsed, err := msg.Read(reader)
	if err != nil {
		sock.readLock.Unlock()
		return SockReadErr{cause: err}
	}
	if !typeParsed {
		sock.readLock.Unlock()
		return SockReadErr{cause: errors.New("no message type")}
	}

	sock.readLock.Unlock()
	return nil
}

// IsConnected implements the webwire.Socket interface
func (sock *Socket) IsConnected() bool {
	sock.lock.Lock()
	connected := sock.connected
	sock.lock.Unlock()
	return connected
}

// RemoteAddr implements the webwire.Socket interface
func (sock *Socket) RemoteAddr() net.Addr {
	sock.lock.Lock()
	if sock.connected {
		addr := sock.conn.RemoteAddr()
		sock.lock.Unlock()
		return addr
	}
	sock.lock.Unlock()
	return nil
}

// Close implements the webwire.Socket interface
func (sock *Socket) Close() error {
	sock.lock.Lock()
	if sock.connected {
		err := sock.conn.Close()
		sock.connected = false
		sock.lock.Unlock()
		return err
	}
	sock.lock.Unlock()
	return nil
}
