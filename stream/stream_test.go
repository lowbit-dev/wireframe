package stream_test

import (
	"bytes"
	"io"
	"testing"

	"lowbit.dev/wireframe/checksum"
	"lowbit.dev/wireframe/frame"
	"lowbit.dev/wireframe/stream"
)

func buildFormat() *frame.Format {
	return &frame.Format{
		Prefix: []byte{0xAA, 0x55},
		Header: []frame.Field{
			frame.Uint8("type"),
			frame.PayloadLength("length"),
		},
		Checksum:   checksum.CRC32IEEE(),
		MaxPayload: 64 * 1024,
	}
}

func encodeFrame(t *testing.T, f *frame.Format, msgType uint8, payload []byte) []byte {
	t.Helper()
	var vals frame.Values
	frame.Set(&vals, "type", msgType)
	out, err := f.Encode(nil, payload, &vals)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return out
}

func TestReaderSingleFrame(t *testing.T) {
	f := buildFormat()
	data := encodeFrame(t, f, 0x01, []byte("hello"))

	sr := stream.NewReader(bytes.NewReader(data), f)
	payload, vals, err := sr.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if string(payload) != "hello" {
		t.Fatalf("payload = %q", payload)
	}
	if vals.Uint8("type") != 0x01 {
		t.Fatalf("type = %d", vals.Uint8("type"))
	}
}

func TestReaderMultipleFrames(t *testing.T) {
	f := buildFormat()
	var buf bytes.Buffer
	for i := uint8(0); i < 5; i++ {
		buf.Write(encodeFrame(t, f, i, []byte("msg")))
	}

	sr := stream.NewReader(&buf, f)
	for i := uint8(0); i < 5; i++ {
		_, vals, err := sr.Next()
		if err != nil {
			t.Fatalf("frame %d: %v", i, err)
		}
		if vals.Uint8("type") != i {
			t.Fatalf("frame %d: type = %d, want %d", i, vals.Uint8("type"), i)
		}
	}

	_, _, err := sr.Next()
	if err != io.EOF {
		t.Fatalf("expected EOF after last frame, got %v", err)
	}
}

func TestReaderFragmented(t *testing.T) {
	f := buildFormat()
	full := encodeFrame(t, f, 0x07, []byte("fragmented"))

	// Feed data one byte at a time using a slow reader.
	sr := stream.NewReader(newSlowReader(full), f)
	payload, vals, err := sr.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if string(payload) != "fragmented" {
		t.Fatalf("payload = %q", payload)
	}
	if vals.Uint8("type") != 0x07 {
		t.Fatalf("type = %#x", vals.Uint8("type"))
	}
}

// slowReader returns one byte per Read call, simulating a slow connection.
type slowReader struct {
	data []byte
	pos  int
}

func newSlowReader(data []byte) *slowReader { return &slowReader{data: data} }

func (r *slowReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	p[0] = r.data[r.pos]
	r.pos++
	return 1, nil
}

// --- TypedReader tests ---

type TypedHdr struct {
	Type   uint8
	Length uint32
}

var typedFmt = frame.NewTypedFormat[TypedHdr](
	frame.WithPrefix[TypedHdr]([]byte{0xAA, 0x55}),
	frame.Uint8Field("type", func(h *TypedHdr) *uint8 { return &h.Type }),
	frame.PayloadLengthField("length", func(h *TypedHdr) *uint32 { return &h.Length }),
	frame.WithChecksum[TypedHdr](checksum.CRC32IEEE()),
	frame.WithMaxPayload[TypedHdr](64*1024),
)

func TestTypedReaderSingleFrame(t *testing.T) {
	hdr := TypedHdr{Type: 0x42}
	data, err := typedFmt.Encode(nil, []byte("typed"), &hdr)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	tr := stream.NewTypedReader(bytes.NewReader(data), typedFmt)
	payload, got, err := tr.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if string(payload) != "typed" {
		t.Fatalf("payload = %q", payload)
	}
	if got.Type != 0x42 {
		t.Fatalf("Type = %#x, want 0x42", got.Type)
	}
}
