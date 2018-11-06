package message

import (
	"fmt"
	"io"

	pld "github.com/qbeon/webwire-go/payload"
)

// WriteMsgSignal writes a named signal message to the given writer closing
// it eventually
func WriteMsgSignal(
	writer io.WriteCloser,
	name []byte,
	payloadEncoding pld.Encoding,
	payloadData []byte,
	safeMode bool,
) error {
	if len(name) > 255 {
		initialErr := fmt.Errorf(
			"unsupported request message name length: %d",
			len(name),
		)
		if err := writer.Close(); err != nil {
			return fmt.Errorf("%s: %s", initialErr, err)
		}
		return initialErr
	}

	// Verify payload data validity in case of UTF16 encoding
	if payloadEncoding == pld.Utf16 && len(payloadData)%2 != 0 {
		initialErr := fmt.Errorf(
			"invalid UTF16 signal payload data length: %d",
			len(payloadData),
		)
		if err := writer.Close(); err != nil {
			return fmt.Errorf("%s: %s", initialErr, err)
		}
		return initialErr
	}

	if safeMode {
		for i := range name {
			char := name[i]
			if char < 32 || char > 126 {
				initialErr := fmt.Errorf(
					"unsupported character in request name: %s",
					string(char),
				)
				if err := writer.Close(); err != nil {
					return fmt.Errorf("%s: %s", initialErr, err)
				}
				return initialErr
			}
		}
	}

	// Determine the message type from the payload encoding type
	sigType := MsgSignalBinary
	switch payloadEncoding {
	case pld.Utf8:
		sigType = MsgSignalUtf8
	case pld.Utf16:
		sigType = MsgSignalUtf16
	}

	// Write message type flag
	if _, err := writer.Write([]byte{sigType}); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			return fmt.Errorf("%s: %s", err, closeErr)
		}
		return err
	}

	// Write name length flag
	if _, err := writer.Write([]byte{byte(len(name))}); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			return fmt.Errorf("%s: %s", err, closeErr)
		}
		return err
	}

	// Write name
	if _, err := writer.Write(name); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			return fmt.Errorf("%s: %s", err, closeErr)
		}
		return err
	}

	// Write header padding byte if the payload requires proper alignment
	if payloadEncoding == pld.Utf16 && len(name)%2 != 0 {
		if _, err := writer.Write([]byte{0}); err != nil {
			if closeErr := writer.Close(); closeErr != nil {
				return fmt.Errorf("%s: %s", err, closeErr)
			}
			return err
		}
	}

	// Write payload
	if _, err := writer.Write(payloadData); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			return fmt.Errorf("%s: %s", err, closeErr)
		}
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}
