package frame_test

import (
	"testing"

	"lowbit.dev/wireframe/checksum"
	"lowbit.dev/wireframe/frame"
)

func buildFormat() *frame.Format {
	return &frame.Format{
		Prefix: []byte{0xAA, 0x55},
		Header: []frame.Field{
			frame.Uint8("version"),
			frame.Uint8("type"),
			frame.Uint8("flags"),
			frame.PayloadLength("length"),
		},
		Checksum:   checksum.CRC32IEEE(),
		MaxPayload: 64 * 1024,
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	f := buildFormat()
	payload := []byte("hello, wireframe")

	var vals frame.Values
	frame.Set(&vals, "version", uint8(1))
	frame.Set(&vals, "type", uint8(0x42))
	frame.Set(&vals, "flags", uint8(0))

	encoded, err := f.Encode(nil, payload, &vals)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, dvals, n, err := f.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if n != len(encoded) {
		t.Fatalf("Decode consumed %d bytes, want %d", n, len(encoded))
	}
	if string(got) != string(payload) {
		t.Fatalf("payload mismatch: got %q, want %q", got, payload)
	}
	if dvals.Uint8("version") != 1 {
		t.Fatalf("version = %d, want 1", dvals.Uint8("version"))
	}
	if dvals.Uint8("type") != 0x42 {
		t.Fatalf("type = %#x, want 0x42", dvals.Uint8("type"))
	}
}

func TestEncodeDecodeNilPayload(t *testing.T) {
	f := buildFormat()

	var vals frame.Values
	frame.Set(&vals, "version", uint8(1))
	frame.Set(&vals, "type", uint8(0x01))
	frame.Set(&vals, "flags", uint8(0))

	encoded, err := f.Encode(nil, nil, &vals)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, _, _, err := f.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty payload, got %v", got)
	}
}

func TestDecodeIncomplete(t *testing.T) {
	f := buildFormat()
	payload := []byte("data")

	var vals frame.Values
	frame.Set(&vals, "version", uint8(1))
	frame.Set(&vals, "type", uint8(1))
	frame.Set(&vals, "flags", uint8(0))

	encoded, _ := f.Encode(nil, payload, &vals)

	// Trim the last byte to force ErrIncomplete.
	_, _, _, err := f.Decode(encoded[:len(encoded)-1])
	if err == nil {
		t.Fatal("expected error for truncated frame")
	}
}

func TestDecodeChecksumError(t *testing.T) {
	f := buildFormat()
	payload := []byte("data")

	var vals frame.Values
	frame.Set(&vals, "version", uint8(1))
	frame.Set(&vals, "type", uint8(1))
	frame.Set(&vals, "flags", uint8(0))

	encoded, _ := f.Encode(nil, payload, &vals)

	// Corrupt the checksum bytes at the end.
	encoded[len(encoded)-1] ^= 0xFF

	_, _, _, err := f.Decode(encoded)
	if err == nil {
		t.Fatal("expected ErrChecksum")
	}
}

func TestDecodeFrameTooLarge(t *testing.T) {
	payload := make([]byte, 100)

	var vals frame.Values
	frame.Set(&vals, "version", uint8(1))
	frame.Set(&vals, "type", uint8(1))
	frame.Set(&vals, "flags", uint8(0))

	// Encode with a format that allows larger payloads.
	bigFormat := &frame.Format{
		Header: []frame.Field{
			frame.PayloadLength("length"),
		},
		MaxPayload: 1000,
	}
	encoded, _ := bigFormat.Encode(nil, payload, &vals)

	// Decode with a format that only allows 10 bytes.
	smallFormat := &frame.Format{
		Header: []frame.Field{
			frame.PayloadLength("length"),
		},
		MaxPayload: 10,
	}
	_, _, _, err := smallFormat.Decode(encoded)
	if err == nil {
		t.Fatal("expected ErrFrameTooLarge")
	}
}

func TestNoChecksumFormat(t *testing.T) {
	f := &frame.Format{
		Header: []frame.Field{
			frame.Uint8("type"),
			frame.PayloadLength("length"),
		},
		MaxPayload: 256,
	}

	payload := []byte("test")
	var vals frame.Values
	frame.Set(&vals, "type", uint8(7))

	encoded, err := f.Encode(nil, payload, &vals)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, dvals, _, err := f.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if string(got) != "test" {
		t.Fatalf("payload = %q", got)
	}
	if dvals.Uint8("type") != 7 {
		t.Fatalf("type = %d", dvals.Uint8("type"))
	}
}

func TestValuesSet(t *testing.T) {
	var vals frame.Values
	frame.Set(&vals, "a", uint8(10))
	frame.Set(&vals, "b", uint16(300))
	frame.Set(&vals, "c", uint32(70000))

	if vals.Uint8("a") != 10 {
		t.Fatalf("Uint8 a = %d", vals.Uint8("a"))
	}
	if vals.Uint16("b") != 300 {
		t.Fatalf("Uint16 b = %d", vals.Uint16("b"))
	}
	if vals.Uint32("c") != 70000 {
		t.Fatalf("Uint32 c = %d", vals.Uint32("c"))
	}

	vals.Reset()
	if vals.Uint8("a") != 0 {
		t.Fatalf("after Reset, Uint8 a = %d", vals.Uint8("a"))
	}
}
