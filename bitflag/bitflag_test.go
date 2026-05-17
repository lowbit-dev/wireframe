package bitflag_test

import (
	"testing"

	"lowbit.dev/wireframe/bitflag"
)

type MsgFlags uint8

const (
	FlagAck      MsgFlags = 1 << 0
	FlagRetry    MsgFlags = 1 << 1
	FlagCompress MsgFlags = 1 << 2
)

func TestHas(t *testing.T) {
	f := bitflag.Of(FlagAck, FlagCompress)

	if !f.Has(FlagAck) {
		t.Error("expected FlagAck to be set")
	}
	if f.Has(FlagRetry) {
		t.Error("expected FlagRetry to be unset")
	}
	if !f.Has(FlagCompress) {
		t.Error("expected FlagCompress to be set")
	}
}

func TestHasAll(t *testing.T) {
	f := bitflag.Of(FlagAck, FlagCompress)

	if !f.HasAll(FlagAck, FlagCompress) {
		t.Error("HasAll should return true for both set flags")
	}
	if f.HasAll(FlagAck, FlagRetry) {
		t.Error("HasAll should return false when one flag is missing")
	}
}

func TestHasAny(t *testing.T) {
	f := bitflag.Of(FlagAck)

	if !f.HasAny(FlagAck, FlagRetry) {
		t.Error("HasAny should return true when at least one flag is set")
	}
	if f.HasAny(FlagRetry, FlagCompress) {
		t.Error("HasAny should return false when no flags are set")
	}
}

func TestSetAndClear(t *testing.T) {
	f := bitflag.Of(FlagAck)
	f.Set(FlagRetry)

	if !f.Has(FlagRetry) {
		t.Error("expected FlagRetry after Set")
	}

	f.Clear(FlagAck)
	if f.Has(FlagAck) {
		t.Error("expected FlagAck to be cleared")
	}
}

func TestValue(t *testing.T) {
	f := bitflag.Of(FlagAck, FlagCompress)
	want := FlagAck | FlagCompress
	if f.Value() != want {
		t.Fatalf("Value() = %#x, want %#x", f.Value(), want)
	}
}

func TestEmpty(t *testing.T) {
	var f bitflag.Flags[MsgFlags]
	if f.Has(FlagAck) {
		t.Error("zero Flags should have nothing set")
	}
}
