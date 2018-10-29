package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
)

// TestClientRequestDisconnected tests sending requests on disconnected clients
func TestClientRequestDisconnected(t *testing.T) {
	// Initialize webwire server given only the request
	server := setupServer(
		t,
		&serverImpl{
			onRequest: func(
				_ context.Context,
				_ wwr.Connection,
				_ wwr.Message,
			) (wwr.Payload, error) {
				return nil, nil
			},
		},
		wwr.ServerOptions{},
	)

	// Initialize client and skip manual connection establishment
	client := newCallbackPoweredClient(
		server.AddressURL(),
		wwrclt.Options{
			DefaultRequestTimeout: 2 * time.Second,
			Autoconnect:           wwrclt.Disabled,
		},
		callbackPoweredClientHooks{},
	)

	// Send request and await reply
	_, err := client.connection.Request(
		context.Background(),
		"",
		wwr.NewPayload(wwr.EncodingBinary, []byte("testdata")),
	)
	require.NoError(t, err)
}
