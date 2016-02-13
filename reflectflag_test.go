package reflectflag

import (
	"bytes"
	"flag"
	"testing"
)

func TestFailNonStruct(t *testing.T) {
	var b bytes.Buffer
	flags := flag.NewFlagSet("testflags", flag.ContinueOnError)
	flags.SetOutput(&b)
	if err := RegisterFlags(flags, 1); err == nil {
		t.Errorf("RegisterFlags returned nil error; expected non-nil")
	}
}

func TestFailMissingTag(t *testing.T) {
	var b bytes.Buffer
	flags := flag.NewFlagSet("testflags", flag.ContinueOnError)
	flags.SetOutput(&b)
	s := struct {
		Field1 string
	}{
		Field1: "field1_value",
	}
	if err := RegisterFlags(flags, s); err == nil {
		t.Errorf("RegisterFlags returned nil error; expected non-nil")
	}
}

func TestTagName(t *testing.T) {
	var b bytes.Buffer
	flags := flag.NewFlagSet("testflags", flag.ContinueOnError)
	flags.SetOutput(&b)
	dflt := struct {
		Field1 string `flag:"field1"`
	}{
		Field1: "field1_value",
	}
	foo := struct {
		Field1 string `foo:"field1"`
	}{
		Field1: "field1_value",
	}
	if err := RegisterFlags(flags, dflt); err != nil {
		t.Errorf("RegisterFlags expected nil error; got %v", err)
	}
	if err := RegisterFlags(flags, foo); err == nil {
		t.Errorf("RegisterFlags returned nil error; expected non-nil")
	}
	if err := RegisterFlags(flags, dflt, TagName("foo")); err == nil {
		t.Errorf("RegisterFlags returned nil error; expected non-nil")
	}
	if err := RegisterFlags(flags, foo, TagName("foo")); err != nil {
		t.Errorf("RegisterFlags expected nil error; got %v", err)
	}
}
