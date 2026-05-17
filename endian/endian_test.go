package endian_test

import (
	"testing"

	"lowbit.dev/wireframe/endian"
)

type MyU32 uint32

func TestRoundTripBE(t *testing.T) {
	buf := make([]byte, 8)

	endian.PutBE(buf, uint8(0xAB))
	if got := endian.ReadBE[uint8](buf); got != 0xAB {
		t.Fatalf("uint8 BE: got %#x", got)
	}

	endian.PutBE(buf, uint16(0xDEAD))
	if got := endian.ReadBE[uint16](buf); got != 0xDEAD {
		t.Fatalf("uint16 BE: got %#x", got)
	}

	endian.PutBE(buf, uint32(0xDEADBEEF))
	if got := endian.ReadBE[uint32](buf); got != 0xDEADBEEF {
		t.Fatalf("uint32 BE: got %#x", got)
	}

	endian.PutBE(buf, uint64(0xDEADBEEFCAFEBABE))
	if got := endian.ReadBE[uint64](buf); got != 0xDEADBEEFCAFEBABE {
		t.Fatalf("uint64 BE: got %#x", got)
	}
}

func TestRoundTripLE(t *testing.T) {
	buf := make([]byte, 8)

	endian.PutLE(buf, uint16(0x1234))
	if got := endian.ReadLE[uint16](buf); got != 0x1234 {
		t.Fatalf("uint16 LE: got %#x", got)
	}

	endian.PutLE(buf, uint32(0xDEADBEEF))
	if got := endian.ReadLE[uint32](buf); got != 0xDEADBEEF {
		t.Fatalf("uint32 LE: got %#x", got)
	}
}

func TestUserDefinedType(t *testing.T) {
	buf := make([]byte, 4)
	endian.PutBE(buf, MyU32(0xCAFEBABE))
	got := endian.ReadBE[MyU32](buf)
	if got != 0xCAFEBABE {
		t.Fatalf("user-defined type BE: got %#x", got)
	}
}

func TestBEByteOrder(t *testing.T) {
	buf := make([]byte, 2)
	endian.PutBE(buf, uint16(0x0102))
	if buf[0] != 0x01 || buf[1] != 0x02 {
		t.Fatalf("big-endian byte order wrong: %v", buf)
	}
}

func TestLEByteOrder(t *testing.T) {
	buf := make([]byte, 2)
	endian.PutLE(buf, uint16(0x0102))
	if buf[0] != 0x02 || buf[1] != 0x01 {
		t.Fatalf("little-endian byte order wrong: %v", buf)
	}
}
