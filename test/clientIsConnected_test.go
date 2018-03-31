package test

import (
	"testing"
	"time"

	webwire "github.com/qbeon/webwire-go"
	webwireClient "github.com/qbeon/webwire-go/client"
)

// TestClientIsConnected tests the client.Status method
func TestClientIsConnected(t *testing.T) {
	// Initialize webwire server given only the request
	server := setupServer(t, webwire.ServerOptions{})

	// Initialize client
	client := webwireClient.NewClient(
		server.Addr().String(),
		webwireClient.Options{
			DefaultRequestTimeout: 2 * time.Second,
		},
	)

	if client.Status() == webwireClient.StatConnected {
		t.Fatal("Expected client to be disconnected before the connection establishment")
	}

	// Connect to the server
	if err := client.Connect(); err != nil {
		t.Fatalf("Couldn't connect the client to the server: %s", err)
	}

	if client.Status() != webwireClient.StatConnected {
		t.Fatal("Expected client to be connected after the connection establishment")
	}

	// Disconnect the client
	client.Close()

	if client.Status() == webwireClient.StatConnected {
		t.Fatal("Expected client to be disconnected after closure")
	}
}
