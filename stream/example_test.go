package stream_test

import (
	"bytes"
	"fmt"

	"lowbit.dev/wireframe/frame"
	"lowbit.dev/wireframe/stream"
)

func ExampleNewReader() {
	f := &frame.Format{
		Header: []frame.Field{
			frame.Uint8("type"),
			frame.PayloadLength("length"),
		},
		MaxPayload: 65535,
	}

	// Encode two frames back-to-back into a buffer.
	var buf bytes.Buffer
	for _, msgType := range []uint8{1, 2} {
		var vals frame.Values
		frame.Set(&vals, "type", msgType)
		data, _ := f.Encode(nil, []byte("ping"), &vals)
		buf.Write(data)
	}

	sr := stream.NewReader(&buf, f)
	for i := 0; i < 2; i++ {
		payload, vals, _ := sr.Next()
		fmt.Printf("%d: %s\n", vals.Uint8("type"), payload)
	}
	// Output:
	// 1: ping
	// 2: ping
}
