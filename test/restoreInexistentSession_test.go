package test

import (
	"context"
	"testing"
	"time"

	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
	"github.com/stretchr/testify/require"
)

// TestRestoreInexistentSession tests the restoration of an inexistent session
func TestRestoreInexistentSession(t *testing.T) {
	// Initialize server
	setup := setupTestServer(
		t,
		&serverImpl{},
		wwr.ServerOptions{},
		nil, // Use the default transport implementation
	)

	// Initialize client

	// Ensure that the last superfluous client is rejected
	client := setup.newClient(
		wwrclt.Options{
			DefaultRequestTimeout: 2 * time.Second,
		},
		nil, // Use the default transport implementation
		testClientHooks{},
	)

	require.NoError(t, client.connection.Connect())

	// Try to restore the session and expect it to fail
	// due to the session being inexistent
	sessionRestorationError := client.connection.RestoreSession(
		context.Background(),
		[]byte("lalala"),
	)
	require.Error(t, sessionRestorationError)
	require.IsType(t, wwr.SessionNotFoundErr{}, sessionRestorationError)
}
