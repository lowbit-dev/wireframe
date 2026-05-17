// Package checksum provides a unified interface for checksum algorithms
// used in binary protocol frames.
//
// All built-in implementations append the checksum to a dst slice
// following the append pattern, making it straightforward to integrate
// into buffer-reuse workflows.
package checksum

import (
	"encoding/binary"
	"hash/adler32"
	"hash/crc32"
	"hash/crc64"
)

// Algorithm is the interface implemented by all checksum algorithms.
type Algorithm interface {
	// Size returns the byte length of the checksum produced by this algorithm.
	Size() int
	// Append computes the checksum of data and appends it to dst.
	Append(dst []byte, data []byte) []byte
	// Verify reports whether the checksum at the end of data is valid.
	// data contains the protected bytes followed immediately by the checksum.
	Verify(data []byte, checksum []byte) bool
}

// -- CRC-32/IEEE --

type crc32IEEE struct{}

// CRC32IEEE returns an Algorithm using the IEEE CRC-32 polynomial.
func CRC32IEEE() Algorithm { return crc32IEEE{} }

func (crc32IEEE) Size() int { return 4 }

func (crc32IEEE) Append(dst []byte, data []byte) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], crc32.ChecksumIEEE(data))
	return append(dst, buf[:]...)
}

func (crc32IEEE) Verify(data []byte, checksum []byte) bool {
	sum := crc32.ChecksumIEEE(data)
	return binary.BigEndian.Uint32(checksum) == sum
}

// -- CRC-64/ISO --

var crc64ISOTable = crc64.MakeTable(crc64.ISO)

type crc64ISO struct{}

// CRC64ISO returns an Algorithm using the ISO CRC-64 polynomial.
func CRC64ISO() Algorithm { return crc64ISO{} }

func (crc64ISO) Size() int { return 8 }

func (crc64ISO) Append(dst []byte, data []byte) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], crc64.Checksum(data, crc64ISOTable))
	return append(dst, buf[:]...)
}

func (crc64ISO) Verify(data []byte, checksum []byte) bool {
	sum := crc64.Checksum(data, crc64ISOTable)
	return binary.BigEndian.Uint64(checksum) == sum
}

// -- XOR-8 --

type xor8 struct{}

// XOR8 returns an Algorithm that XORs all bytes together, producing a
// single-byte checksum. Cheap, but only suitable for simple error detection.
func XOR8() Algorithm { return xor8{} }

func (xor8) Size() int { return 1 }

func (xor8) Append(dst []byte, data []byte) []byte {
	var v byte
	for _, b := range data {
		v ^= b
	}
	return append(dst, v)
}

func (xor8) Verify(data []byte, checksum []byte) bool {
	var v byte
	for _, b := range data {
		v ^= b
	}
	return checksum[0] == v
}

// -- CRC-16 --

// crc16 uses the CRC-16/ANSI (CRC-16/IBM) polynomial 0x8005 with
// initial value 0x0000, which is the variant used by Modbus RTU.
type crc16 struct{}

// CRC16 returns an Algorithm using the CRC-16/ANSI (Modbus) polynomial.
func CRC16() Algorithm { return crc16{} }

func (crc16) Size() int { return 2 }

func (c crc16) Append(dst []byte, data []byte) []byte {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], computeCRC16(data))
	return append(dst, buf[:]...)
}

func (c crc16) Verify(data []byte, checksum []byte) bool {
	return binary.LittleEndian.Uint16(checksum) == computeCRC16(data)
}

func computeCRC16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// -- Adler-32 --

type adler32alg struct{}

// Adler32 returns an Algorithm using the Adler-32 checksum.
func Adler32() Algorithm { return adler32alg{} }

func (adler32alg) Size() int { return 4 }

func (adler32alg) Append(dst []byte, data []byte) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], adler32.Checksum(data))
	return append(dst, buf[:]...)
}

func (adler32alg) Verify(data []byte, checksum []byte) bool {
	return binary.BigEndian.Uint32(checksum) == adler32.Checksum(data)
}
