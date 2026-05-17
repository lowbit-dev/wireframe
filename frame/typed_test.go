package frame_test

import (
	"testing"

	"lowbit.dev/wireframe/checksum"
	"lowbit.dev/wireframe/frame"
)

type MsgHeader struct {
	Version uint8
	Type    uint8
	Flags   uint8
	Length  uint32
}

var msgFormat = frame.NewTypedFormat[MsgHeader](
	frame.WithPrefix[MsgHeader]([]byte{0xAA, 0x55}),
	frame.Uint8Field("version", func(h *MsgHeader) *uint8 { return &h.Version }),
	frame.Uint8Field("type", func(h *MsgHeader) *uint8 { return &h.Type }),
	frame.Uint8Field("flags", func(h *MsgHeader) *uint8 { return &h.Flags }),
	frame.PayloadLengthField("length", func(h *MsgHeader) *uint32 { return &h.Length }),
	frame.WithChecksum[MsgHeader](checksum.CRC32IEEE()),
	frame.WithMaxPayload[MsgHeader](64*1024),
)

func TestTypedEncodeDecodeRoundTrip(t *testing.T) {
	payload := []byte("typed wireframe")

	hdr := MsgHeader{Version: 2, Type: 0x10, Flags: 0x01}
	encoded, err := msgFormat.Encode(nil, payload, &hdr)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var decoded MsgHeader
	got, n, err := msgFormat.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if n != len(encoded) {
		t.Fatalf("consumed %d bytes, want %d", n, len(encoded))
	}
	if string(got) != string(payload) {
		t.Fatalf("payload = %q, want %q", got, payload)
	}
	if decoded.Version != 2 {
		t.Fatalf("Version = %d, want 2", decoded.Version)
	}
	if decoded.Type != 0x10 {
		t.Fatalf("Type = %#x, want 0x10", decoded.Type)
	}
	if decoded.Flags != 0x01 {
		t.Fatalf("Flags = %#x, want 0x01", decoded.Flags)
	}
	if decoded.Length != uint32(len(payload)) {
		t.Fatalf("Length = %d, want %d", decoded.Length, len(payload))
	}
}

func TestTypedFormat_Format(t *testing.T) {
	f := msgFormat.Format()
	if f == nil {
		t.Fatal("Format() returned nil")
	}
	if f.MaxPayload != 64*1024 {
		t.Fatalf("MaxPayload = %d", f.MaxPayload)
	}
}
