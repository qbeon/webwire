package main

import (
	"context"

	"github.com/valyala/fasthttp"

	wwr "github.com/qbeon/webwire-go"
)

// BenchmarkServer implements the webwire.ServerImplementation interface
type BenchmarkServer struct{}

// OnOptions implements the webwire.ServerImplementation interface
func (srv *BenchmarkServer) OnOptions(_ *fasthttp.RequestCtx) {}

// OnSignal implements the webwire.ServerImplementation interface
func (srv *BenchmarkServer) OnSignal(
	_ context.Context,
	_ wwr.Connection,
	_ wwr.Message,
) {
}

// OnClientConnected implements the webwire.ServerImplementation interface
func (srv *BenchmarkServer) OnClientConnected(conn wwr.Connection) {}

// OnClientDisconnected implements the webwire.ServerImplementation interface
func (srv *BenchmarkServer) OnClientDisconnected(
	conn wwr.Connection,
	reason error,
) {
}

// OnRequest implements the webwire.ServerImplementation interface.
// Returns the received message back to the client
func (srv *BenchmarkServer) OnRequest(
	ctx context.Context,
	_ wwr.Connection,
	msg wwr.Message,
) (response wwr.Payload, err error) {
	// Reply to the request using the same data and encoding
	return wwr.Payload{
		Encoding: msg.PayloadEncoding(),
		Data:     msg.Payload(),
	}, nil
}
