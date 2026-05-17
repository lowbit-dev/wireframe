// Package msgproto demonstrates a complete binary messaging protocol built
// with wireframe. It uses every layer of the stack: frame for the envelope,
// schema for the payload layout, stream for incremental reading, and bitflag
// for the flag field.
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
//	go run lowbit.dev/wireframe/examples/msgproto
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

// Header is the per-frame header struct used with TypedFormat.
type Header struct {
	Version uint8
	Type    uint8
	Flags   uint8
	Length  uint32
}

var protoFormat = frame.NewTypedFormat[Header](
	frame.WithPrefix[Header]([]byte{0xAA, 0x55}),
	frame.Uint8Field("version", func(h *Header) *uint8 { return &h.Version }),
	frame.Uint8Field("type", func(h *Header) *uint8 { return &h.Type }),
	frame.Uint8Field("flags", func(h *Header) *uint8 { return &h.Flags }),
	frame.PayloadLengthField("length", func(h *Header) *uint32 { return &h.Length }),
	frame.WithChecksum[Header](checksum.CRC32IEEE()),
	frame.WithMaxPayload[Header](64*1024),
)

// --- Payload schemas -----------------------------------------------------

// DataMsg carries a sequence number and arbitrary data.
type DataMsg struct {
	Seq  uint16
	Data []byte
}

var dataSchema = schema.NewTyped[DataMsg](
	schema.Uint16BEField("seq", func(m *DataMsg) *uint16 { return &m.Seq }),
	schema.VarBytesField("data", func(m *DataMsg) *[]byte { return &m.Data }),
)

// AckMsg acknowledges a DataMsg by its sequence number.
type AckMsg struct {
	Seq uint16
}

var ackSchema = schema.NewTyped[AckMsg](
	schema.Uint16BEField("seq", func(a *AckMsg) *uint16 { return &a.Seq }),
)

// --- Encode helpers ------------------------------------------------------

func encodePing(priority bool) ([]byte, error) {
	flags := bitflag.Of[MsgFlags]()
	if priority {
		flags.Set(FlagPriority)
	}
	hdr := Header{
		Version: protoVersion,
		Type:    uint8(MsgPing),
		Flags:   uint8(flags.Value()),
	}
	return protoFormat.Encode(nil, nil, &hdr)
}

func encodeData(seq uint16, data []byte) ([]byte, error) {
	payload, err := dataSchema.Encode(nil, &DataMsg{Seq: seq, Data: data})
	if err != nil {
		return nil, err
	}
	hdr := Header{
		Version: protoVersion,
		Type:    uint8(MsgData),
	}
	return protoFormat.Encode(nil, payload, &hdr)
}

func encodeAck(seq uint16) ([]byte, error) {
	payload, err := ackSchema.Encode(nil, &AckMsg{Seq: seq})
	if err != nil {
		return nil, err
	}
	hdr := Header{
		Version: protoVersion,
		Type:    uint8(MsgDataAck),
	}
	return protoFormat.Encode(nil, payload, &hdr)
}

// --- Main ----------------------------------------------------------------

func main() {
	// Write a stream of frames to a buffer: one ping, two data messages,
	// two acks. A real application would write to a net.Conn or serial port.
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

	// Read back via the typed stream reader. Each call to Next returns the
	// payload and a fully populated Header without any string key lookups.
	sr := stream.NewTypedReader(&buf, protoFormat)

	for {
		payload, hdr, err := sr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		flags := bitflag.Of(MsgFlags(hdr.Flags))

		switch MsgType(hdr.Type) {
		case MsgPing:
			fmt.Printf("PING    version=%d priority=%v\n",
				hdr.Version, flags.Has(FlagPriority))

		case MsgData:
			var msg DataMsg
			if _, err := dataSchema.Decode(payload, &msg); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("DATA    seq=%d data=%q\n", msg.Seq, msg.Data)

		case MsgDataAck:
			var ack AckMsg
			if _, err := ackSchema.Decode(payload, &ack); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("ACK     seq=%d\n", ack.Seq)

		default:
			fmt.Printf("UNKNOWN type=0x%02x len=%d\n", hdr.Type, len(payload))
		}
	}
}
