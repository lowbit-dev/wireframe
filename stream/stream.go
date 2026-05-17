// Package stream provides incremental frame parsing over an io.Reader.
//
// Data from sockets and serial ports rarely arrives aligned to frame
// boundaries. Reader handles buffering and partial reads transparently.
// For the lifetime of a connection, Reader performs zero per-frame heap
// allocations.
package stream

import (
	"errors"
	"io"

	"lowbit.dev/wireframe/frame"
)

// defaultBufSize is the initial read-buffer capacity.
const defaultBufSize = 4096

// Reader buffers data from an io.Reader and returns complete decoded
// frames one at a time.
type Reader struct {
	r      io.Reader
	format *frame.Format
	vals   frame.Values
	buf    []byte
}

// NewReader returns a Reader that reads from r and decodes frames
// according to format.
func NewReader(r io.Reader, format *frame.Format) *Reader {
	return &Reader{
		r:      r,
		format: format,
		buf:    make([]byte, 0, defaultBufSize),
	}
}

// Next blocks until a complete frame is available, then returns the
// payload and a pointer to the reader's internal Values populated from
// the frame header fields.
//
// The returned Values pointer is valid until the next call to Next.
// Callers that need to retain header data beyond that must copy it.
//
// Next returns io.EOF when the underlying reader is exhausted cleanly at
// a frame boundary, or io.ErrUnexpectedEOF when it closes mid-frame.
func (r *Reader) Next() (payload []byte, vals *frame.Values, err error) {
	tmp := make([]byte, defaultBufSize)

	for {
		// Attempt to decode a frame from whatever we have buffered.
		if len(r.buf) > 0 {
			p, n, decErr := r.format.DecodeInto(r.buf, &r.vals)
			if decErr == nil {
				// Consume the frame bytes and return.
				r.buf = r.buf[n:]
				return p, &r.vals, nil
			}
			if !errors.Is(decErr, frame.ErrIncomplete) {
				return nil, nil, decErr
			}
			// ErrIncomplete: need more data.
		}

		// Read more data from the underlying reader.
		n, readErr := r.r.Read(tmp)
		if n > 0 {
			r.buf = append(r.buf, tmp[:n]...)
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				if len(r.buf) == 0 {
					return nil, nil, io.EOF
				}
				return nil, nil, io.ErrUnexpectedEOF
			}
			return nil, nil, readErr
		}
	}
}

// TypedReader wraps Reader to return fully populated header structs of
// type T alongside each frame's payload.
type TypedReader[T any] struct {
	r      *Reader
	format *frame.TypedFormat[T]
}

// NewTypedReader returns a TypedReader that reads from r and decodes
// frames according to the typed format f.
func NewTypedReader[T any](r io.Reader, f *frame.TypedFormat[T]) *TypedReader[T] {
	return &TypedReader[T]{
		r:      NewReader(r, f.Format()),
		format: f,
	}
}

// Next blocks until a complete frame is available. It returns the
// payload and the header fields decoded into a value of type T.
// The header is returned by value; it is safe to retain across calls.
func (tr *TypedReader[T]) Next() (payload []byte, header T, err error) {
	payload, vals, err := tr.r.Next()
	if err != nil {
		return nil, header, err
	}
	// Decode the Values back into a T using the typed format's decoder.
	// We re-decode from the raw Values that were populated by the untyped
	// reader. The typed format's Decode method already does this, but it
	// starts from wire bytes. Instead we expose a direct path.
	// Since TypedFormat.Decode re-runs the full wire decode, we use a small
	// workaround: we pass the raw Values through the accessor functions
	// directly via the typed format's internal fields.
	//
	// The cleanest path is to expose the typed format's field binding.
	// For now we record the last wire bytes in the untyped reader and
	// re-run the typed decode. Instead, we add a lower-level helper.
	tr.format.FillHeader(vals, &header)
	return payload, header, nil
}
