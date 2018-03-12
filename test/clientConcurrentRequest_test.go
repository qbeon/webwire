package test

import (
	"context"
	"os"
	"testing"
	"time"

	webwire "github.com/qbeon/webwire-go"
	webwireClient "github.com/qbeon/webwire-go/client"
)

// TestClientConcurrentRequest verifies concurrent calling of client.Request
// is properly synchronized and doesn't cause any data race
func TestClientConcurrentRequest(t *testing.T) {
	var concurrentAccessors uint32 = 16
	finished := NewPending(concurrentAccessors*2, 2*time.Second, true)

	// Initialize webwire server
	_, addr := setupServer(
		t,
		webwire.Hooks{
			OnRequest: func(_ context.Context) (webwire.Payload, error) {
				finished.Done()
				return webwire.Payload{}, nil
			},
		},
	)

	// Initialize client
	client := webwireClient.NewClient(
		addr,
		webwireClient.Hooks{},
		5*time.Second,
		os.Stdout,
		os.Stderr,
	)
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("Couldn't connect: %s", err)
	}

	sendRequest := func() {
		defer finished.Done()
		if _, err := client.Request(
			"sample",
			webwire.Payload{Data: []byte("samplepayload")},
		); err != nil {
			t.Errorf("Request failed: %s", err)
		}
	}

	for i := uint32(0); i < concurrentAccessors; i++ {
		go sendRequest()
	}

	if err := finished.Wait(); err != nil {
		t.Fatal("Expectation timed out")
	}
}
