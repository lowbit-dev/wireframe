package frame_test

import (
	"fmt"

	"lowbit.dev/wireframe/checksum"
	"lowbit.dev/wireframe/frame"
)

// ExHdr is the wire header for the TypedFormat example.
type ExHdr struct {
	Version uint8
	Type    uint8
	Length  uint32
}

var exFmt = frame.NewTypedFormat[ExHdr](
	frame.WithPrefix[ExHdr]([]byte{0xAA, 0x55}),
	frame.Uint8Field("version", func(h *ExHdr) *uint8 { return &h.Version }),
	frame.Uint8Field("type", func(h *ExHdr) *uint8 { return &h.Type }),
	frame.PayloadLengthField("length", func(h *ExHdr) *uint32 { return &h.Length }),
	frame.WithChecksum[ExHdr](checksum.CRC32IEEE()),
	frame.WithMaxPayload[ExHdr](65535),
)

func ExampleNewTypedFormat() {
	hdr := ExHdr{Version: 1, Type: 0x42}
	encoded, _ := exFmt.Encode(nil, []byte("world"), &hdr)

	var decoded ExHdr
	payload, _, _ := exFmt.Decode(encoded, &decoded)
	fmt.Println(string(payload))
	fmt.Println(decoded.Version)
	fmt.Println(decoded.Type)   // 0x42 = 66
	fmt.Println(decoded.Length) // len("world") = 5
	// Output:
	// world
	// 1
	// 66
	// 5
}
