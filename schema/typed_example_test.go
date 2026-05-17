package schema_test

import (
	"fmt"

	"lowbit.dev/wireframe/schema"
)

// Request is a typed payload for the TypedSchema example.
type Request struct {
	Op   uint8
	Seq  uint16
	Data []byte
}

var reqSchema = schema.NewTyped[Request](
	schema.Uint8Field("op", func(r *Request) *uint8 { return &r.Op }),
	schema.Uint16BEField("seq", func(r *Request) *uint16 { return &r.Seq }),
	schema.VarBytesField("data", func(r *Request) *[]byte { return &r.Data }),
)

func ExampleNewTyped() {
	msg := Request{Op: 5, Seq: 100, Data: []byte("hello")}
	encoded, _ := reqSchema.Encode(nil, &msg)

	var decoded Request
	reqSchema.Decode(encoded, &decoded)
	fmt.Println(decoded.Op)
	fmt.Println(decoded.Seq)
	fmt.Println(string(decoded.Data))
	// Output:
	// 5
	// 100
	// hello
}
