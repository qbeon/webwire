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

// TestCustomSessKeyGen tests custom session key generators
func TestCustomSessKeyGen(t *testing.T) {
	expectedSessionKey := "customkey123"

	// Initialize webwire server
	setup := setupTestServer(
		t,
		&serverImpl{
			onRequest: func(
				_ context.Context,
				conn wwr.Connection,
				_ wwr.Message,
			) (wwr.Payload, error) {
				// Try to create a new session
				err := conn.CreateSession(nil)
				assert.NoError(t, err)
				if err != nil {
					return wwr.Payload{}, err
				}

				key := conn.SessionKey()
				assert.Equal(t, expectedSessionKey, key)

				// Return the key of the newly created session
				// (use default binary encoding)
				return wwr.Payload{Data: []byte(key)}, nil
			},
		},
		wwr.ServerOptions{
			SessionKeyGenerator: &sessionKeyGen{
				generate: func() string {
					return expectedSessionKey
				},
			},
		},
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
	defer client.connection.Close()

	// Send authentication request and await reply
	_, err := client.connection.Request(
		context.Background(),
		[]byte("login"),
		wwr.Payload{Data: []byte("testdata")},
	)
	require.NoError(t, err)

	require.Equal(t, expectedSessionKey, client.connection.Session().Key)
}
