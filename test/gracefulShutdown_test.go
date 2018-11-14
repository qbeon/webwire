package test

import (
	"context"
	"testing"
	"time"

	tmdwg "github.com/qbeon/tmdwg-go"
	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGracefulShutdown tests the ability of the server to delay shutdown
// until all requests and signals are processed
// and reject incoming connections and requests
// while ignoring incoming signals
//
// SIGNAL:       |-->||||||||||----------------- (must finish)
// REQUEST:      ||-->||||||||||---------------- (must finish)
// SRV SHUTDWN:  |||----->||||||---------------- (must await req and sig)
// LATE CONN:    |||------->|------------------- (must be rejected)
// LATE REQ:     ||||------->|------------------ (must be rejected)
func TestGracefulShutdown(t *testing.T) {
	expectedReqReply := []byte("i_finished")
	handlerExecutionDuration := 100 * time.Millisecond
	maxTestDuration := handlerExecutionDuration * 4
	firstReqAndSigSent := tmdwg.NewTimedWaitGroup(2, maxTestDuration)
	serverShuttingDown := tmdwg.NewTimedWaitGroup(1, maxTestDuration)
	handlersFinished := tmdwg.NewTimedWaitGroup(2, maxTestDuration)
	serverShutDown := tmdwg.NewTimedWaitGroup(
		1,
		2*time.Second,
	)

	// Initialize webwire server
	setup := setupTestServer(
		t,
		&serverImpl{
			onSignal: func(
				_ context.Context,
				_ wwr.Connection,
				msg wwr.Message,
			) {
				if string(msg.Name()) == "1" {
					firstReqAndSigSent.Progress(1)
				}
				// Sleep after the first signal was marked as done
				time.Sleep(handlerExecutionDuration)
				handlersFinished.Progress(1)
			},
			onRequest: func(
				_ context.Context,
				_ wwr.Connection,
				msg wwr.Message,
			) (wwr.Payload, error) {
				if string(msg.Name()) == "1" {
					firstReqAndSigSent.Progress(1)
				}
				time.Sleep(handlerExecutionDuration)
				return wwr.Payload{Data: expectedReqReply}, nil
			},
		},
		wwr.ServerOptions{},
		nil, // Use the default transport implementation
	)

	// Initialize different clients for the signal,
	// the request and the late request and conn
	// to avoid serializing them because every client
	// is handled in a separate goroutine
	cltOpts := wwrclt.Options{
		DefaultRequestTimeout: 5 * time.Second,
		Autoconnect:           wwr.Disabled,
	}
	clientSig := setup.newClient(
		cltOpts,
		nil, // Use the default transport implementation
		testClientHooks{},
	)
	clientReq := setup.newClient(
		cltOpts,
		nil, // Use the default transport implementation
		testClientHooks{},
	)
	clientLateReq := setup.newClient(
		cltOpts,
		nil, // Use the default transport implementation
		testClientHooks{},
	)

	require.NoError(t, clientSig.connection.Connect())
	require.NoError(t, clientReq.connection.Connect())
	require.NoError(t, clientLateReq.connection.Connect())

	// Disable autoconnect for the late client to enable immediate errors
	clientLateConn := setup.newClient(
		wwrclt.Options{
			Autoconnect: wwr.Disabled,
		},
		nil, // Use the default transport implementation
		testClientHooks{},
	)

	// Send signal and request in another parallel goroutine
	// to avoid blocking the main test goroutine when awaiting the request reply
	go func() {
		// (SIGNAL)
		assert.NoError(t, clientSig.connection.Signal(
			context.Background(),
			[]byte("1"),
			wwr.Payload{Data: []byte("test")},
		))

		// (REQUEST)
		rep, err := clientReq.connection.Request(
			context.Background(),
			[]byte("1"),
			wwr.Payload{Data: []byte("test")},
		)
		assert.NoError(t, err)
		assert.Equal(t, string(rep.Payload()), string(expectedReqReply))
		handlersFinished.Progress(1)
		rep.Close()
	}()

	// Request server shutdown in another parallel goroutine
	// to avoid blocking the main test goroutine when waiting
	// for the server to shut down
	go func() {
		// Wait for the signal and request to arrive and get handled,
		// then request the shutdown
		assert.NoError(t,
			firstReqAndSigSent.Wait(),
			"First request and signal were not sent within %s",
			handlerExecutionDuration,
		)

		// (SRV SHUTDWN)
		serverShuttingDown.Progress(1)
		setup.Server.Shutdown()
		serverShutDown.Progress(1)
	}()

	// Wait for the server to start shutting down and fire late requests
	// and late connection in another parallel goroutine
	// to avoid blocking the main test goroutine when performing them
	go func() {
		// Wait for the server to start shutting down
		assert.NoError(t,
			serverShuttingDown.Wait(),
			"Server not shutting down after %s",
			maxTestDuration,
		)

		// Verify connection establishment during shutdown (LATE CONN)
		assert.Error(t,
			clientLateConn.connection.Connect(),
			"Expected late connection to be rejected, "+
				"though it still was accepted",
		)

		// Verify request rejection during shutdown (LATE REQ)
		_, lateReqErr := clientLateReq.connection.Request(
			context.Background(),
			nil,
			wwr.Payload{Data: []byte("test")},
		)
		switch err := lateReqErr.(type) {
		case wwr.ServerShutdownErr:
			break
		case wwr.RequestErr:
			t.Errorf("Expected special server shutdown error, "+
				"got regular request error: %s",
				err,
			)
		default:
			t.Errorf("Expected request during shutdown to be rejected " +
				"with special error type",
			)
		}
	}()

	// Await server shutdown, timeout if necessary
	require.NoError(t,
		serverShutDown.Wait(),
		"Expected server to shut down within %s",
		maxTestDuration,
	)

	// Expect both the signal and the request to have completed properly
	require.NoError(t,
		handlersFinished.Wait(),
		"Expected signal and request to have finished processing",
	)
}
