package schema_test

import (
	"testing"

	"lowbit.dev/wireframe/schema"
)

func TestUntypedRoundTrip(t *testing.T) {
	s := schema.New(
		schema.Uint8("command"),
		schema.Uint16BE("data_length"),
		schema.VarBytes("data", "data_length"),
	)

	vals := schema.NewValues()
	schema.Set(vals, "command", uint8(0x01))
	schema.Set(vals, "data", []byte("hello"))

	encoded, err := s.Encode(nil, vals)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded := schema.NewValues()
	n, err := s.Decode(encoded, decoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if n != len(encoded) {
		t.Fatalf("consumed %d, want %d", n, len(encoded))
	}
	if decoded.Uint8("command") != 0x01 {
		t.Fatalf("command = %d", decoded.Uint8("command"))
	}
	if string(decoded.Bytes("data")) != "hello" {
		t.Fatalf("data = %q", decoded.Bytes("data"))
	}
}

func TestBoolField(t *testing.T) {
	s := schema.New(schema.Bool("flag"))

	vals := schema.NewValues()
	schema.Set(vals, "flag", true)

	encoded, err := s.Encode(nil, vals)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(encoded) != 1 || encoded[0] != 1 {
		t.Fatalf("encoded bool = %v", encoded)
	}

	decoded := schema.NewValues()
	_, err = s.Decode(encoded, decoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	rv := decoded.Uint8("flag")
	if rv == 0 {
		t.Fatal("expected true flag")
	}
}

func TestVarintField(t *testing.T) {
	s := schema.New(schema.Varint("seq"))

	vals := schema.NewValues()
	schema.Set(vals, "seq", uint32(300))

	encoded, err := s.Encode(nil, vals)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// 300 needs 2 bytes as a varint.
	if len(encoded) != 2 {
		t.Fatalf("expected 2 bytes for varint 300, got %d", len(encoded))
	}

	decoded := schema.NewValues()
	_, err = s.Decode(encoded, decoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if decoded.Uint32("seq") != 300 {
		t.Fatalf("seq = %d, want 300", decoded.Uint32("seq"))
	}
}

func TestFixedBytesField(t *testing.T) {
	s := schema.New(schema.Bytes("id", 4))

	vals := schema.NewValues()
	schema.Set(vals, "id", []byte{0xDE, 0xAD, 0xBE, 0xEF})

	encoded, _ := s.Encode(nil, vals)
	decoded := schema.NewValues()
	s.Decode(encoded, decoded)

	id := decoded.Bytes("id")
	if id[0] != 0xDE || id[3] != 0xEF {
		t.Fatalf("id = %v", id)
	}
}

// --- TypedSchema tests ---

type Payload struct {
	Command uint8
	Seq     uint16
	Data    []byte
}

var payloadSchema = schema.NewTyped(
	schema.Uint8Field("command", func(p *Payload) *uint8 { return &p.Command }),
	schema.Uint16BEField("seq", func(p *Payload) *uint16 { return &p.Seq }),
	schema.VarBytesField("data", func(p *Payload) *[]byte { return &p.Data }),
)

func TestTypedSchemaRoundTrip(t *testing.T) {
	msg := Payload{Command: 0x03, Seq: 42, Data: []byte("world")}

	encoded, err := payloadSchema.Encode(nil, &msg)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var decoded Payload
	n, err := payloadSchema.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if n != len(encoded) {
		t.Fatalf("consumed %d, want %d", n, len(encoded))
	}
	if decoded.Command != 0x03 {
		t.Fatalf("Command = %d", decoded.Command)
	}
	if decoded.Seq != 42 {
		t.Fatalf("Seq = %d", decoded.Seq)
	}
	if string(decoded.Data) != "world" {
		t.Fatalf("Data = %q", decoded.Data)
	}
}
