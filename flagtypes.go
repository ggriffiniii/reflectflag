package reflectflag

import (
	"flag"
	"fmt"
	"strconv"
	"time"
)

type boolValue bool

func newBoolValue(b interface{}) flag.Getter {
	rb := boolValue(b.(bool))
	return &rb
}

func (b *boolValue) Set(val string) error {
	v, err := strconv.ParseBool(val)
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() interface{} { return bool(*b) }

func (b *boolValue) String() string { return fmt.Sprintf("%v", *b) }

func (b *boolValue) IsBoolFlag() bool { return true }

type intValue int

func newIntValue(i interface{}) flag.Getter {
	ri := intValue(i.(int))
	return &ri
}

func (i *intValue) Set(val string) error {
	v, err := strconv.ParseInt(val, 0, 64)
	*i = intValue(v)
	return err
}

func (i *intValue) Get() interface{} { return int(*i) }

func (i *intValue) String() string { return fmt.Sprintf("%v", *i) }

type int32Value int32

func newInt32Value(i interface{}) flag.Getter {
	ri := int32Value(i.(int32))
	return &ri
}

func (i *int32Value) Set(val string) error {
	v, err := strconv.ParseInt(val, 0, 32)
	*i = int32Value(v)
	return err
}

func (i *int32Value) Get() interface{} { return int32(*i) }

func (i *int32Value) String() string { return fmt.Sprintf("%v", *i) }

type int64Value int64

func newInt64Value(i interface{}) flag.Getter {
	ri := int64Value(i.(int64))
	return &ri
}

func (i *int64Value) Set(val string) error {
	v, err := strconv.ParseInt(val, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() interface{} { return int64(*i) }

func (i *int64Value) String() string { return fmt.Sprintf("%v", *i) }

type uintValue uint

func newUintValue(i interface{}) flag.Getter {
	ri := uintValue(i.(uint))
	return &ri
}

func (i *uintValue) Set(val string) error {
	v, err := strconv.ParseUint(val, 0, 64)
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() interface{} { return uint(*i) }

func (i *uintValue) String() string { return fmt.Sprintf("%v", *i) }

type uint32Value uint32

func newUint32Value(i interface{}) flag.Getter {
	ri := uint32Value(i.(uint32))
	return &ri
}

func (i *uint32Value) Set(val string) error {
	v, err := strconv.ParseUint(val, 0, 32)
	*i = uint32Value(v)
	return err
}

func (i *uint32Value) Get() interface{} { return uint32(*i) }

func (i *uint32Value) String() string { return fmt.Sprintf("%v", *i) }

type uint64Value uint64

func newUint64Value(i interface{}) flag.Getter {
	ri := uint64Value(i.(uint64))
	return &ri
}

func (i *uint64Value) Set(val string) error {
	v, err := strconv.ParseUint(val, 0, 64)
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() interface{} { return uint64(*i) }

func (i *uint64Value) String() string { return fmt.Sprintf("%v", *i) }

type float32Value float32

func newFloat32Value(f interface{}) flag.Getter {
	rf := float32Value(f.(float32))
	return &rf
}

func (f *float32Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 32)
	*f = float32Value(v)
	return err
}
func (f *float32Value) Get() interface{} { return float32(*f) }
func (f *float32Value) String() string   { return fmt.Sprintf("%v", *f) }

type float64Value float64

func newFloat64Value(f interface{}) flag.Getter {
	rf := float64Value(f.(float64))
	return &rf
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float64Value(v)
	return err
}
func (f *float64Value) Get() interface{} { return float64(*f) }
func (f *float64Value) String() string   { return fmt.Sprintf("%v", *f) }

type stringValue string

func newStringValue(s interface{}) flag.Getter {
	rs := stringValue(s.(string))
	return &rs
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return fmt.Sprintf("%s", *s) }

type durationValue time.Duration

func newDurationValue(d interface{}) flag.Getter {
	rd := durationValue(d.(time.Duration))
	return &rd
}

func (d *durationValue) Set(val string) error {
	v, err := time.ParseDuration(val)
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() interface{} { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }
