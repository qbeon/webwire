package test

import (
	"context"
	"testing"
	"time"

	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClientRequestRegisterOnTimeout verifies the request register
// of the client is correctly updated when the request times out
func TestClientRequestRegisterOnTimeout(t *testing.T) {
	var connection wwrclt.Client

	// Initialize webwire server given only the request
	setup := setupTestServer(
		t,
		&serverImpl{
			onRequest: func(
				_ context.Context,
				_ wwr.Connection,
				_ wwr.Message,
			) (wwr.Payload, error) {
				// Verify pending requests
				assert.Equal(t, 1, connection.PendingRequests())

				// Wait until the request times out
				time.Sleep(300 * time.Millisecond)
				return wwr.Payload{}, nil
			},
		},
		wwr.ServerOptions{},
		nil, // Use the default transport implementation
	)

	// Initialize client
	client := setup.newClient(
		wwrclt.Options{
			DefaultRequestTimeout: 2 * time.Second,
		},
		nil, // Use the default transport implementation
		testClientHooks{},
	)
	connection = client.connection

	// Connect the client to the server
	require.NoError(t, client.connection.Connect())

	// Verify pending requests
	require.Equal(t, 0, client.connection.PendingRequests())

	// Send request and await reply
	contextWithDeadline, cancel := context.WithTimeout(
		context.Background(),
		200*time.Millisecond,
	)
	defer cancel()
	_, reqErr := client.connection.Request(
		contextWithDeadline,
		nil,
		wwr.Payload{Data: []byte("t")},
	)
	require.Error(t, reqErr)
	require.IsType(t, wwr.DeadlineExceededErr{}, reqErr)
	require.True(t, wwr.IsTimeoutErr(reqErr))
	require.False(t, wwr.IsCanceledErr(reqErr))

	// Verify pending requests
	require.Equal(t, 0, client.connection.PendingRequests())
}
