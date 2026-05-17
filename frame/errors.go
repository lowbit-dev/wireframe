package frame

import "errors"

// Sentinel errors returned by Format.Encode and Format.Decode.
var (
	// ErrIncomplete is returned when the source slice does not contain
	// enough bytes to decode a complete frame.
	ErrIncomplete = errors.New("frame: incomplete")

	// ErrChecksum is returned when checksum verification fails.
	ErrChecksum = errors.New("frame: checksum mismatch")

	// ErrFrameTooLarge is returned when the decoded payload length
	// exceeds Format.MaxPayload.
	ErrFrameTooLarge = errors.New("frame: payload too large")

	// ErrInvalidFormat is returned when a Format is misconfigured.
	ErrInvalidFormat = errors.New("frame: invalid format")
)
