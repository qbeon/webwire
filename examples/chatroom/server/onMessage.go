package main

import (
	"context"
	"log"

	wwr "github.com/qbeon/webwire-go"
)

// onMessage handles incoming chat messages.
// The server tries to identify the client by its session
// and falls back to "anonymous" if the client has no ongoing session
func onMessage(ctx context.Context) (wwr.Payload, error) {
	msg := ctx.Value(wwr.Msg).(wwr.Message)
	client := msg.Client

	msgStr, err := msg.Payload.Utf8()
	if err != nil {
		log.Printf(
			"Received invalid message from %s, couldn't convert payload to UTF8: %s",
			client.RemoteAddr(),
			err,
		)
		return wwr.Payload{}, nil
	}

	log.Printf(
		"Received message from %s: '%s' (%d)",
		client.RemoteAddr(),
		msgStr,
		len(msg.Payload.Data),
	)

	name := "Anonymous"
	// Try to read the name from the session
	if msg.Client.Session != nil {
		sessionInfo := msg.Client.Session.Info.(map[string]string)
		name = sessionInfo["username"]
	}

	broadcastMessage(name, msgStr)

	return wwr.Payload{}, nil
}
