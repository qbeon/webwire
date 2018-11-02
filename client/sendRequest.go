package client

import (
	"context"
	"fmt"
	"time"

	webwire "github.com/qbeon/webwire-go"
	msg "github.com/qbeon/webwire-go/message"
)

func (clt *client) sendRequest(
	ctx context.Context,
	messageType byte,
	name []byte,
	payload webwire.Payload,
	timeout time.Duration,
) (webwire.Payload, error) {
	// Require either a name or a payload or both
	if len(name) < 1 && (payload == nil || len(payload.Data()) < 1) {
		return nil, webwire.NewProtocolErr(
			fmt.Errorf("Invalid request, request message requires " +
				"either a name, a payload or both but is missing both",
			),
		)
	}

	payloadEncoding := webwire.EncodingBinary
	var payloadData []byte
	if payload != nil {
		payloadEncoding = payload.Encoding()
		payloadData = payload.Data()
	}

	// Compose a message and register it
	request := clt.requestManager.Create(timeout)
	reqIdentifier := request.Identifier()
	msg := msg.NewRequestMessage(
		reqIdentifier,
		name,
		payloadEncoding,
		payloadData,
	)

	// Return an error if the request was already prematurely canceled
	// or already exceeded the user-defined deadline for its completion
	select {
	case <-ctx.Done():
		err := webwire.TranslateContextError(ctx.Err())
		clt.requestManager.Fail(reqIdentifier, err)
		return nil, err
	default:
	}

	// Send request
	if err := clt.conn.Write(msg); err != nil {
		return nil, webwire.NewReqTransErr(err)
	}

	clt.heartbeat.reset()

	// Block until request either times out or a response is received
	return request.AwaitReply(ctx)
}
