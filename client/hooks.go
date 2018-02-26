package client

import webwire "github.com/qbeon/webwire-go"

// Hooks represents all callback hook functions
type Hooks struct {
	// OnServerSignal is an optional callback.
	// It's invoked when the webwire client receives a signal from the server
	OnServerSignal func([]byte)

	// OnSessionCreated is an optional callback.
	// It's invoked when the webwire client receives a new session
	OnSessionCreated func(*webwire.Session)

	// OnSessionClosed is an optional callback.
	// It's invoked when the clients session was closed
	// either by the server or by himself
	OnSessionClosed func()
}

// SetDefaults sets undefined required hooks
func (hooks *Hooks) SetDefaults() {
	if hooks.OnServerSignal == nil {
		hooks.OnServerSignal = func(_ []byte) {}
	}

	if hooks.OnSessionCreated == nil {
		hooks.OnSessionCreated = func(_ *webwire.Session) {}
	}

	if hooks.OnSessionClosed == nil {
		hooks.OnSessionClosed = func() {}
	}
}
