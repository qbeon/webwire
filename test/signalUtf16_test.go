package test

import (
	"context"
	"sync"
	"testing"
	"time"

	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignalUtf16 tests client-side signals with UTF16 encoded payloads
func TestSignalUtf16(t *testing.T) {
	testPayload := wwr.Payload{
		Encoding: wwr.EncodingUtf16,
		Data:     []byte{00, 115, 00, 97, 00, 109, 00, 112, 00, 108, 00, 101},
	}
	signalArrived := sync.WaitGroup{}
	signalArrived.Add(1)

	// Initialize webwire server given only the signal handler
	setup := SetupTestServer(
		t,
		&ServerImpl{
			Signal: func(
				_ context.Context,
				_ wwr.Connection,
				msg wwr.Message,
			) {
				assert.Equal(t, wwr.EncodingUtf16, msg.PayloadEncoding())
				assert.Equal(t, testPayload.Data, msg.Payload())

				// Synchronize, notify signal arrival
				signalArrived.Done()
			},
		},
		wwr.ServerOptions{},
		nil, // Use the default transport implementation
	)

	// Initialize client
	client := setup.NewClient(
		wwrclt.Options{
			DefaultRequestTimeout: 2 * time.Second,
		},
		nil, // Use the default transport implementation
		TestClientHooks{},
	)

	require.NoError(t, client.Connection.Connect())

	// Send signal
	require.NoError(t, client.Connection.Signal(
		context.Background(),
		nil,
		wwr.Payload{
			Encoding: wwr.EncodingUtf16,
			Data: []byte{
				00, 115, 00, 97, 00, 109,
				00, 112, 00, 108, 00, 101,
			},
		},
	))

	// Synchronize, await signal arrival
	signalArrived.Wait()
}