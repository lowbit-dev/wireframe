package stream_test

import (
	"bytes"
	"fmt"

	"lowbit.dev/wireframe/checksum"
	"lowbit.dev/wireframe/frame"
	"lowbit.dev/wireframe/stream"
)

// EgHeader is the wire header for the TypedReader example.
type EgHeader struct {
	Type   uint8
	Length uint32
}

var egFmt = frame.NewTypedFormat[EgHeader](
	frame.Uint8Field("type", func(h *EgHeader) *uint8 { return &h.Type }),
	frame.PayloadLengthField("length", func(h *EgHeader) *uint32 { return &h.Length }),
	frame.WithChecksum[EgHeader](checksum.CRC32IEEE()),
	frame.WithMaxPayload[EgHeader](65535),
)

func ExampleNewTypedReader() {
	hdr := EgHeader{Type: 7}
	data, _ := egFmt.Encode(nil, []byte("typed"), &hdr)

	tr := stream.NewTypedReader(bytes.NewReader(data), egFmt)
	payload, got, _ := tr.Next()
	fmt.Println(string(payload))
	fmt.Println(got.Type)
	// Output:
	// typed
	// 7
}
