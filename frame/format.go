package frame

import (
	"bytes"
	"fmt"

	"lowbit.dev/wireframe/checksum"
)

// Format describes the structure of a binary frame. It is responsible for
// packet boundaries and integrity — not payload content.
//
// A frame is laid out as:
//
//	[Prefix] [Header fields...] [Payload] [Checksum] [Suffix]
type Format struct {
	// Prefix is an optional fixed byte sequence at the start of every
	// frame (e.g. a magic number or SOF marker).
	Prefix []byte

	// Header is the ordered list of fixed header fields. A PayloadLength
	// field must be included when the format uses a length-prefixed payload.
	Header []Field

	// Checksum algorithm to use. nil disables checksum.
	Checksum checksum.Algorithm

	// Suffix is an optional fixed byte sequence appended after the checksum.
	Suffix []byte

	// MaxPayload is the maximum allowed payload size in bytes. It must be
	// greater than zero; Format.Validate returns ErrInvalidFormat otherwise.
	MaxPayload int

	// payloadLengthIdx is the index of the PayloadLength field in Header,
	// or -1 if there is none. Populated by Validate.
	payloadLengthIdx int

	// fixedHeaderSize is the total byte size of all fixed-width header
	// fields. Populated by Validate; -1 if any field has variable size.
	fixedHeaderSize int

	// validated is set to true after Validate has completed successfully.
	validated bool
}

// Validate checks the Format for consistency and pre-computes derived
// fields. Encode and Decode call this lazily on first use; callers may
// call it explicitly to catch configuration errors early.
func (f *Format) Validate() error {
	if f.MaxPayload <= 0 {
		return fmt.Errorf("%w: MaxPayload must be > 0", ErrInvalidFormat)
	}

	f.payloadLengthIdx = -1
	f.fixedHeaderSize = 0

	for i, field := range f.Header {
		if _, ok := field.(payloadLengthField); ok {
			if f.payloadLengthIdx != -1 {
				return fmt.Errorf("%w: multiple PayloadLength fields", ErrInvalidFormat)
			}
			f.payloadLengthIdx = i
		}
		sz := field.Size()
		if sz < 0 {
			f.fixedHeaderSize = -1
		} else if f.fixedHeaderSize >= 0 {
			f.fixedHeaderSize += sz
		}
	}

	f.validated = true
	return nil
}

// checksumSize returns the byte length of the checksum, or 0.
func (f *Format) checksumSize() int {
	if f.Checksum == nil {
		return 0
	}
	return f.Checksum.Size()
}

// Encode encodes payload and the header values in vals into a frame,
// appending the result to dst. The caller must set all header fields in
// vals before calling Encode; PayloadLength is set automatically.
//
// Returns the extended dst slice or an error.
func (f *Format) Encode(dst []byte, payload []byte, vals *Values) ([]byte, error) {
	if err := f.ensureValidated(); err != nil {
		return nil, err
	}
	if len(payload) > f.MaxPayload {
		return nil, ErrFrameTooLarge
	}

	// Set the payload length field automatically.
	if f.payloadLengthIdx >= 0 {
		Set(vals, f.Header[f.payloadLengthIdx].Name(), uint32(len(payload)))
	}

	// The checksum covers prefix + header + payload.
	start := len(dst)

	dst = append(dst, f.Prefix...)

	// Encode each header field. For fixed-width fields we grow dst and
	// encode directly into the tail. For variable-width (varint) fields
	// we use a small scratch buffer.
	var scratch [10]byte
	for _, field := range f.Header {
		sz := field.Size()
		if sz >= 0 {
			// Fixed-width: grow dst, encode into tail.
			dst = append(dst, scratch[:sz]...)
			n, err := field.Encode(dst[len(dst)-sz:], vals)
			if err != nil {
				return nil, err
			}
			_ = n // n == sz for fixed-width fields
		} else {
			// Variable-width: encode into scratch, then append.
			n, err := field.Encode(scratch[:], vals)
			if err != nil {
				return nil, err
			}
			dst = append(dst, scratch[:n]...)
		}
	}

	dst = append(dst, payload...)

	// Compute and append checksum over prefix+header+payload.
	if f.Checksum != nil {
		dst = f.Checksum.Append(dst, dst[start:])
	}

	dst = append(dst, f.Suffix...)

	return dst, nil
}

// Decode decodes a single frame from src. It returns the payload, the
// populated header Values, the number of bytes consumed from src, and
// any error.
//
// ErrIncomplete is returned when src is too short; the caller should
// buffer more data and retry.
func (f *Format) Decode(src []byte) (payload []byte, vals *Values, n int, err error) {
	vals = NewValues()
	payload, n, err = f.DecodeInto(src, vals)
	if err != nil {
		return nil, nil, 0, err
	}
	return payload, vals, n, nil
}

// DecodeInto decodes a single frame from src into the provided vals.
// vals is reset before use. This is the allocation-free decode path used
// by stream.Reader.
func (f *Format) DecodeInto(src []byte, vals *Values) (payload []byte, n int, err error) {
	if err := f.ensureValidated(); err != nil {
		return nil, 0, err
	}

	vals.Reset()
	pos := 0

	// Check and consume prefix.
	if len(f.Prefix) > 0 {
		if len(src)-pos < len(f.Prefix) {
			return nil, 0, ErrIncomplete
		}
		if !bytes.Equal(src[pos:pos+len(f.Prefix)], f.Prefix) {
			return nil, 0, fmt.Errorf("%w: prefix mismatch", ErrInvalidFormat)
		}
		pos += len(f.Prefix)
	}

	// Decode header fields.
	for _, field := range f.Header {
		sz := field.Size()
		minBytes := sz
		if sz < 0 {
			minBytes = 1
		}
		if len(src)-pos < minBytes {
			return nil, 0, ErrIncomplete
		}
		consumed, decErr := field.Decode(src[pos:], vals)
		if decErr != nil {
			return nil, 0, decErr
		}
		pos += consumed
	}

	// Read payload length.
	var payloadLen int
	if f.payloadLengthIdx >= 0 {
		payloadLen = int(vals.Uint32(f.Header[f.payloadLengthIdx].Name()))
		if payloadLen > f.MaxPayload {
			return nil, 0, ErrFrameTooLarge
		}
	}

	checksumSize := f.checksumSize()
	suffixLen := len(f.Suffix)

	// Ensure the rest of the frame is available.
	needed := pos + payloadLen + checksumSize + suffixLen
	if len(src) < needed {
		return nil, 0, ErrIncomplete
	}

	payload = src[pos : pos+payloadLen]
	pos += payloadLen

	// Verify checksum if present.
	if f.Checksum != nil {
		protected := src[0:pos]
		sum := src[pos : pos+checksumSize]
		if !f.Checksum.Verify(protected, sum) {
			return nil, 0, ErrChecksum
		}
		pos += checksumSize
	}

	// Consume suffix.
	pos += suffixLen

	return payload, pos, nil
}

// ensureValidated calls Validate if it has not been called yet.
func (f *Format) ensureValidated() error {
	if f.validated {
		return nil
	}
	return f.Validate()
}
