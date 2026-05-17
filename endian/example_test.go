package endian_test

import (
	"fmt"

	"lowbit.dev/wireframe/endian"
)

func ExamplePutBE() {
	buf := make([]byte, 4)
	endian.PutBE(buf, uint32(0xDEADBEEF))
	fmt.Printf("%X\n", buf)
	// Output:
	// DEADBEEF
}

func ExampleReadBE() {
	buf := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	v := endian.ReadBE[uint32](buf)
	fmt.Printf("%X\n", v)
	// Output:
	// DEADBEEF
}

func ExamplePutLE() {
	buf := make([]byte, 2)
	endian.PutLE(buf, uint16(0x0102))
	// Little-endian stores the low byte first.
	fmt.Printf("%02X %02X\n", buf[0], buf[1])
	// Output:
	// 02 01
}

func ExampleReadLE() {
	buf := []byte{0x02, 0x01}
	v := endian.ReadLE[uint16](buf)
	fmt.Printf("%04X\n", v)
	// Output:
	// 0102
}
