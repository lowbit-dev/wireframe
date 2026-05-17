package varint_test

import (
	"testing"

	"lowbit.dev/wireframe/varint"
)

func TestEncodeDecodeU64(t *testing.T) {
	cases := []uint64{0, 1, 127, 128, 255, 1 << 14, 1 << 21, 1<<63 - 1, ^uint64(0)}

	buf := make([]byte, 10)
	for _, v := range cases {
		n := varint.EncodeU64(buf, v)
		got, m, err := varint.DecodeU64(buf[:n])
		if err != nil {
			t.Fatalf("DecodeU64(%d): %v", v, err)
		}
		if got != v {
			t.Fatalf("round-trip u64 %d: got %d", v, got)
		}
		if m != n {
			t.Fatalf("consumed %d bytes, encoded %d", m, n)
		}
	}
}

func TestEncodeDecodeI64(t *testing.T) {
	cases := []int64{0, 1, -1, 127, -127, 1 << 14, -(1 << 14), 1<<62 - 1, -(1 << 62)}

	buf := make([]byte, 10)
	for _, v := range cases {
		n := varint.EncodeI64(buf, v)
		got, m, err := varint.DecodeI64(buf[:n])
		if err != nil {
			t.Fatalf("DecodeI64(%d): %v", v, err)
		}
		if got != v {
			t.Fatalf("round-trip i64 %d: got %d", v, got)
		}
		if m != n {
			t.Fatalf("consumed %d bytes, encoded %d", m, n)
		}
	}
}

func TestSizeU64(t *testing.T) {
	buf := make([]byte, 10)
	cases := []uint64{0, 127, 128, 16383, 16384, ^uint64(0)}
	for _, v := range cases {
		n := varint.EncodeU64(buf, v)
		s := varint.SizeU64(v)
		if s != n {
			t.Fatalf("SizeU64(%d) = %d, want %d", v, s, n)
		}
	}
}

func TestSizeI64(t *testing.T) {
	buf := make([]byte, 10)
	cases := []int64{0, 1, -1, 63, -64, 64, -65, 1<<62 - 1, -(1 << 62)}
	for _, v := range cases {
		n := varint.EncodeI64(buf, v)
		s := varint.SizeI64(v)
		if s != n {
			t.Fatalf("SizeI64(%d) = %d, want %d", v, s, n)
		}
	}
}

func TestDecodeU64Incomplete(t *testing.T) {
	// A varint with the high bit set on the last byte signals more bytes needed.
	_, _, err := varint.DecodeU64([]byte{0x80})
	if err == nil {
		t.Fatal("expected error for incomplete varint")
	}
}
