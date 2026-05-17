// Package msgproto_values demonstrates the same protocol as examples/msgproto
// using the untyped Values-based API instead of TypedFormat and TypedSchema.
//
// Both examples produce identical wire bytes. The difference is entirely in
// how the Go side interacts with the codec:
//
//   - msgproto:        TypedFormat[T] + TypedSchema[T] — struct fields, compiler-checked
//   - msgproto_values: frame.Format + schema.Schema   — string keys, runtime-checked
//
// The Values API is the right choice when the format is not known at compile
// time, when fields vary per connection, or when you want to keep the format
// declaration and the struct types independent.
//
// Wire layout:
//
//	AA 55 | version(u8) | type(u8) | flags(u8) | length(u32LE) | payload | CRC32
//
// Message types:
//
//	0x01  Ping     — no payload; FlagPriority may be set
//	0x02  Data     — payload: seq(u16BE) + data(var bytes)
//	0x03  DataAck  — payload: seq(u16BE)
//
// Run with:
//
//	go run lowbit.dev/wireframe/examples/msgproto_values
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"lowbit.dev/wireframe/bitflag"
	"lowbit.dev/wireframe/checksum"
	"lowbit.dev/wireframe/frame"
	"lowbit.dev/wireframe/schema"
	"lowbit.dev/wireframe/stream"
)

// --- Protocol constants --------------------------------------------------

const protoVersion = 1

// MsgType identifies the message kind carried in the frame header.
type MsgType uint8

const (
	MsgPing    MsgType = 0x01
	MsgData    MsgType = 0x02
	MsgDataAck MsgType = 0x03
)

// MsgFlags is the flag field in the frame header.
type MsgFlags uint8

const (
	FlagPriority   MsgFlags = 1 << 0
	FlagCompressed MsgFlags = 1 << 1
)

// --- Frame format --------------------------------------------------------

var protoFormat = &frame.Format{
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

// --- Payload schemas -----------------------------------------------------

// dataSchema describes the Data message payload: seq(u16BE) + var bytes.
// VarBytesAuto embeds a 2-byte length prefix automatically, matching the
// wire layout produced by VarBytesField in the typed example.
var dataSchema = schema.New(
	schema.Uint16BE("seq"),
	schema.VarBytesAuto("data"),
)

// ackSchema describes the DataAck message payload: seq(u16BE).
var ackSchema = schema.New(
	schema.Uint16BE("seq"),
)

// --- Encode helpers ------------------------------------------------------

func encodePing(priority bool) ([]byte, error) {
	flags := bitflag.Of[MsgFlags]()
	if priority {
		flags.Set(FlagPriority)
	}
	var vals frame.Values
	frame.Set(&vals, "version", uint8(protoVersion))
	frame.Set(&vals, "type", uint8(MsgPing))
	frame.Set(&vals, "flags", uint8(flags.Value()))
	return protoFormat.Encode(nil, nil, &vals)
}

func encodeData(seq uint16, data []byte) ([]byte, error) {
	var svals schema.Values
	schema.Set(&svals, "seq", seq)
	schema.Set(&svals, "data", data)
	payload, err := dataSchema.Encode(nil, &svals)
	if err != nil {
		return nil, err
	}

	var vals frame.Values
	frame.Set(&vals, "version", uint8(protoVersion))
	frame.Set(&vals, "type", uint8(MsgData))
	frame.Set(&vals, "flags", uint8(0))
	return protoFormat.Encode(nil, payload, &vals)
}

func encodeAck(seq uint16) ([]byte, error) {
	var svals schema.Values
	schema.Set(&svals, "seq", seq)
	payload, err := ackSchema.Encode(nil, &svals)
	if err != nil {
		return nil, err
	}

	var vals frame.Values
	frame.Set(&vals, "version", uint8(protoVersion))
	frame.Set(&vals, "type", uint8(MsgDataAck))
	frame.Set(&vals, "flags", uint8(0))
	return protoFormat.Encode(nil, payload, &vals)
}

// --- Main ----------------------------------------------------------------

func main() {
	var buf bytes.Buffer

	ping, err := encodePing(true)
	if err != nil {
		log.Fatal(err)
	}
	buf.Write(ping)

	for seq := uint16(1); seq <= 2; seq++ {
		msg := fmt.Sprintf("hello from message %d", seq)
		data, err := encodeData(seq, []byte(msg))
		if err != nil {
			log.Fatal(err)
		}
		buf.Write(data)
	}

	for seq := uint16(1); seq <= 2; seq++ {
		ack, err := encodeAck(seq)
		if err != nil {
			log.Fatal(err)
		}
		buf.Write(ack)
	}

	fmt.Printf("encoded %d bytes into buffer\n\n", buf.Len())

	// Read back via the untyped stream reader. vals is a pointer to the
	// reader's internal Values — valid until the next call to Next.
	sr := stream.NewReader(&buf, protoFormat)

	for {
		payload, vals, err := sr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		flags := bitflag.Of(MsgFlags(vals.Uint8("flags")))

		switch MsgType(vals.Uint8("type")) {
		case MsgPing:
			fmt.Printf("PING    version=%d priority=%v\n",
				vals.Uint8("version"), flags.Has(FlagPriority))

		case MsgData:
			var svals schema.Values
			if _, err := dataSchema.Decode(payload, &svals); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("DATA    seq=%d data=%q\n",
				svals.Uint16("seq"), svals.Bytes("data"))

		case MsgDataAck:
			var svals schema.Values
			if _, err := ackSchema.Decode(payload, &svals); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("ACK     seq=%d\n", svals.Uint16("seq"))

		default:
			fmt.Printf("UNKNOWN type=0x%02x len=%d\n",
				vals.Uint8("type"), len(payload))
		}
	}
}
