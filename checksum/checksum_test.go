package checksum_test

import (
	"testing"

	"lowbit.dev/wireframe/checksum"
)

var algorithms = []struct {
	name string
	alg  checksum.Algorithm
}{
	{"CRC32IEEE", checksum.CRC32IEEE()},
	{"CRC64ISO", checksum.CRC64ISO()},
	{"XOR8", checksum.XOR8()},
	{"CRC16", checksum.CRC16()},
	{"Adler32", checksum.Adler32()},
}

func TestSizeMatchesAppend(t *testing.T) {
	data := []byte("hello, wireframe")
	for _, tc := range algorithms {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.alg.Append(nil, data)
			if len(out) != tc.alg.Size() {
				t.Fatalf("Size() = %d, but Append produced %d bytes", tc.alg.Size(), len(out))
			}
		})
	}
}

func TestVerifyValid(t *testing.T) {
	data := []byte("hello, wireframe")
	for _, tc := range algorithms {
		t.Run(tc.name, func(t *testing.T) {
			sum := tc.alg.Append(nil, data)
			if !tc.alg.Verify(data, sum) {
				t.Fatal("Verify returned false for valid checksum")
			}
		})
	}
}

func TestVerifyCorrupt(t *testing.T) {
	data := []byte("hello, wireframe")
	for _, tc := range algorithms {
		t.Run(tc.name, func(t *testing.T) {
			sum := tc.alg.Append(nil, data)
			// Flip one bit in the checksum.
			sum[0] ^= 0xFF
			if tc.alg.Verify(data, sum) {
				t.Fatal("Verify returned true for corrupted checksum")
			}
		})
	}
}

func TestCRC16ModbusKnownVector(t *testing.T) {
	// Modbus RTU known-good frame: 11 03 00 6B 00 03 -> CRC 76 87
	// Address is 0x11 (device 17). CRC bytes are little-endian.
	data := []byte{0x11, 0x03, 0x00, 0x6B, 0x00, 0x03}
	sum := checksum.CRC16().Append(nil, data)
	if sum[0] != 0x76 || sum[1] != 0x87 {
		t.Fatalf("CRC16 known vector: got %02x %02x, want 76 87", sum[0], sum[1])
	}
}
