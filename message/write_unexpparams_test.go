package message_test

import (
	"testing"

	"github.com/qbeon/webwire-go/message"
	pld "github.com/qbeon/webwire-go/payload"
	"github.com/stretchr/testify/require"
)

/****************************************************************\
	Constructors - unexpected parameters (panics)
\****************************************************************/

// TestWriteMsgReqNoNameNoPayload tests WriteMsgRequest
// without both the name and the payload
func TestWriteMsgReqNoNameNoPayload(t *testing.T) {
	writer := &testWriter{}
	require.Error(t, message.WriteMsgRequest(
		writer,
		genRndMsgIdentifier(),
		nil,
		pld.Binary,
		nil,
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgReqNameTooLong tests WriteMsgRequest with a too long name
func TestWriteMsgReqNameTooLong(t *testing.T) {
	tooLongNamelength := 256
	nameBuf := make([]byte, tooLongNamelength)
	for i := 0; i < tooLongNamelength; i++ {
		nameBuf[i] = 'a'
	}

	writer := &testWriter{}
	require.Error(t, message.WriteMsgRequest(
		writer,
		genRndMsgIdentifier(),
		nameBuf,
		0,
		nil,
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgReqInvalidCharsetBelowAscii32 tests WriteMsgRequest
// with an invalid character input below the ASCII 7 bit 32nd character
func TestWriteMsgReqInvalidCharsetBelowAscii32(t *testing.T) {
	// Generate invalid name using a character
	// below the ASCII 7 bit 32nd character
	invalidNameBytes := make([]byte, 1)
	invalidNameBytes[0] = byte(31)

	writer := &testWriter{}
	require.Error(t, message.WriteMsgRequest(
		writer,
		genRndMsgIdentifier(),
		invalidNameBytes,
		0,
		nil,
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgReqInvalidCharsetAboveAscii126 tests WriteMsgRequest
// with an invalid character input above the ASCII 7 bit 126th character
func TestWriteMsgReqInvalidCharsetAboveAscii126(t *testing.T) {
	// Generate invalid name using a character
	// above the ASCII 7 bit 126th character
	invalidNameBytes := make([]byte, 1)
	invalidNameBytes[0] = byte(127)

	writer := &testWriter{}
	require.Error(t, message.WriteMsgRequest(
		writer,
		genRndMsgIdentifier(),
		invalidNameBytes,
		0,
		nil,
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgSigNameTooLong tests WriteMsgSignal with a too long name
func TestWriteMsgSigNameTooLong(t *testing.T) {
	tooLongNamelength := 256
	nameBuf := make([]byte, tooLongNamelength)
	for i := 0; i < tooLongNamelength; i++ {
		nameBuf[i] = 'a'
	}

	writer := &testWriter{}
	require.Error(t, message.WriteMsgSignal(writer, nameBuf, 0, nil, true))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgSigInvalidCharsetBelowAscii32 tests WriteMsgSignal
// with an invalid character input below the ASCII 7 bit 32nd character
func TestWriteMsgSigInvalidCharsetBelowAscii32(t *testing.T) {
	// Generate invalid name using a character
	// below the ASCII 7 bit 32nd character
	invalidNameBytes := make([]byte, 1)
	invalidNameBytes[0] = byte(31)

	writer := &testWriter{}
	require.Error(t, message.WriteMsgSignal(
		writer,
		invalidNameBytes,
		0,
		nil,
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgSigInvalidCharsetAboveAscii126 tests WriteMsgSignal
// with an invalid character input above ASCII 7 bit 126th character
func TestWriteMsgSigInvalidCharsetAboveAscii126(t *testing.T) {
	// Generate invalid name using a character
	// above the ASCII 7 bit 126th character
	invalidNameBytes := make([]byte, 1)
	invalidNameBytes[0] = byte(127)

	writer := &testWriter{}
	require.Error(t, message.WriteMsgSignal(
		writer,
		invalidNameBytes,
		0,
		nil,
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgSpecialRequestReplyInvalidType tests
// WriteMsgSpecialRequestReply with non-special reply message types
func TestWriteMsgSpecialRequestReplyInvalidType(t *testing.T) {
	allTypes := []byte{
		message.MsgReplyError,
		message.MsgNotifySessionCreated,
		message.MsgNotifySessionClosed,
		message.MsgRequestCloseSession,
		message.MsgRequestRestoreSession,
		message.MsgSignalBinary,
		message.MsgSignalUtf8,
		message.MsgSignalUtf16,
		message.MsgRequestBinary,
		message.MsgRequestUtf8,
		message.MsgRequestUtf16,
		message.MsgReplyBinary,
		message.MsgReplyUtf8,
		message.MsgReplyUtf16,
	}

	for _, tp := range allTypes {
		writer := &testWriter{}
		require.Error(t, message.WriteMsgSpecialRequestReply(
			writer,
			tp,
			genRndMsgIdentifier(),
		))
		require.True(t, writer.closed)
		require.Nil(t, writer.buf)
	}
}

// TestWriteMsgReplyErrorNoCode tests WriteMsgReplyError
// with no error code which is invalid.
func TestWriteMsgReplyErrorNoCode(t *testing.T) {
	writer := &testWriter{}
	require.Error(t, message.WriteMsgReplyError(
		writer,
		genRndMsgIdentifier(),
		[]byte(""),
		[]byte("sample error message"),
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgReplyErrorCodeTooLong tests WriteMsgReplyError
// with an error code that's surpassing the 255 character limit.
func TestWriteMsgReplyErrorCodeTooLong(t *testing.T) {
	tooLongCode := make([]byte, 256)
	for i := 0; i < 256; i++ {
		tooLongCode[i] = 'a'
	}

	writer := &testWriter{}
	require.Error(t, message.WriteMsgReplyError(
		writer,
		genRndMsgIdentifier(),
		tooLongCode,
		[]byte("sample error message"),
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgReplyErrorCodeCharsetBelowAscii32 tests WriteMsgReplyError
// with an invalid character input below the ASCII 7 bit 32nd character
func TestWriteMsgReplyErrorCodeCharsetBelowAscii32(t *testing.T) {
	// Generate invalid error code using a character
	// below the ASCII 7 bit 32nd character
	invalidCodeBytes := make([]byte, 1)
	invalidCodeBytes[0] = byte(31)

	writer := &testWriter{}
	require.Error(t, message.WriteMsgReplyError(
		writer,
		genRndMsgIdentifier(),
		invalidCodeBytes,
		[]byte("sample error message"),
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}

// TestWriteMsgReplyErrorCodeCharsetAboveAscii126 tests
// WriteMsgReplyError with an invalid character input
// above the ASCII 7 bit 126th character
func TestWriteMsgReplyErrorCodeCharsetAboveAscii126(t *testing.T) {
	// Generate invalid error code using a character
	// above the ASCII 7 bit 126th character
	invalidCodeBytes := make([]byte, 1)
	invalidCodeBytes[0] = byte(127)

	writer := &testWriter{}
	require.Error(t, message.WriteMsgReplyError(
		writer,
		genRndMsgIdentifier(),
		invalidCodeBytes,
		[]byte("sample error message"),
		true,
	))
	require.True(t, writer.closed)
	require.Nil(t, writer.buf)
}
