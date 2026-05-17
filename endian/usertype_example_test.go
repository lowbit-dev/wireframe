package endian_test

import (
	"fmt"

	"lowbit.dev/wireframe/endian"
)

// SessionID is a user-defined type with an unsigned integer base.
// It satisfies the Unsigned constraint without any explicit casting.
type SessionID uint32

func ExamplePutBE_userDefinedType() {
	buf := make([]byte, 4)
	endian.PutBE(buf, SessionID(42))
	got := endian.ReadBE[SessionID](buf)
	fmt.Println(got)
	// Output:
	// 42
}
