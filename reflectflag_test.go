package reflectflag

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func ptrTo(x interface{}) interface{} {
	v := reflect.New(reflect.TypeOf(x))
	p := v.Elem()
	p.Set(reflect.ValueOf(x))
	return v.Interface()
}

// deepEqual is a "simple" implementation of a recusive comparison, where
// pointers are considered equal if the dereferenced contents are equal. This
// is only intended to work for the types used in this test and is not a
// general purpose utility.
func deepEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	if !va.IsValid() || !vb.IsValid() {
		return va.IsValid() == vb.IsValid()
	}
	if va.Type() != vb.Type() {
		return false
	}
	switch va.Kind() {
	case reflect.Ptr:
		return deepEqual(va.Elem().Interface(), vb.Elem().Interface())
	case reflect.Struct:
		typ := va.Type()
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).PkgPath != "" {
				continue  // skip unexported fields
			}
			if !deepEqual(va.Field(i).Interface(), vb.Field(i).Interface()) {
				return false
			}
		}
		return true
	}
	return a == b
}

type nested struct {
	Nested nested1
}
type nested1 struct {
	Nested nested2
}
type nested2 struct {
	Field string `flag:"nested_flag"`
}

type customType int32
func newCustomType(s string) *customType {
	x := customType(0)
	switch s {
		case "foo": x = customType(1)
		case "bar": x = customType(2)
		case "baz": x = customType(3)
	}
	return &x
}

type customFlag struct {
	t *customType
}
func newCustomFlag(c interface{}) flag.Getter {
	return &customFlag{t: c.(*customType)}
}

func (c *customFlag) String() string {
	switch int32(*c.t) {
	case 1: return "foo"
	case 2: return "bar"
	case 3: return "baz"
	}
	return "invalid"
}

func (c *customFlag) Set(s string) error {
	switch s {
	case "foo":
		*c.t = customType(1)
		return nil
	case "bar":
		*c.t = customType(2)
		return nil
	case "baz":
		*c.t = customType(3)
		return nil
	}
	return errors.New("unable to set customType")
}

func (c *customFlag) Get() interface{} {
	return c.t
}

type testCase struct {
	desc            string
	testStruct      interface{}
	wantRegisterErr error
	wantPreParse    map[string]interface{}
	args            []string
	wantParseErr    error
	wantStruct	interface{}
	wantLoadErr error
	opts []Option
}

func runTestCase(tc testCase) error {
	var b bytes.Buffer
	flags := flag.NewFlagSet("testflags", flag.ContinueOnError)
	flags.SetOutput(&b)

	err := RegisterFlags(flags, tc.testStruct, tc.opts...)
	if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tc.wantRegisterErr) {
		return fmt.Errorf("unexpected return from RegisterFlags; got %v want %v", err, tc.wantRegisterErr)
	}
	if err != nil {
		return nil
	}
	for name, value := range tc.wantPreParse {
		f := flags.Lookup(name)
		if f == nil {
			return fmt.Errorf("Missing flag. Expected to find %s", name)
		}
		v := f.Value.(flag.Getter).Get()
		if !deepEqual(v, value) {
			return fmt.Errorf("Mismatched default flag value for flag %s; got %v want %v", name, v, value)
		}
	}
	err = flags.Parse(tc.args)
	if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tc.wantParseErr) {
		return fmt.Errorf("unexpected return from flags.Parse; got %v want %v", err, tc.wantParseErr)
	}
	if err != nil {
		return nil
	}
	out := reflect.New(reflect.TypeOf(tc.testStruct))
	err = LoadFromFlags(flags, out.Interface(), tc.opts...)
	if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tc.wantLoadErr) {
		return fmt.Errorf("unexpected return from LoadFromFlags; got %v want %v", err, tc.wantLoadErr)
	}
	if err != nil {
		return nil
	}
	result := out.Elem().Interface()
	if !deepEqual(result, tc.wantStruct) {
		return fmt.Errorf("unexpected output from LoadFromFlags; got %#v want %#v", result, tc.wantStruct)
	}
	return nil
}

type allTypesStruct struct {
			BoolFlag        bool           `flag:"bool"`
			BoolPtrFlag     *bool          `flag:"bool_ptr"`
			IntFlag         int            `flag:"int"`
			IntPtrFlag      *int           `flag:"int_ptr"`
			Int32Flag       int32          `flag:"int32"`
			Int32PtrFlag    *int32         `flag:"int32_ptr"`
			Int64Flag       int64          `flag:"int64"`
			Int64PtrFlag    *int64         `flag:"int64_ptr"`
			UintFlag        uint           `flag:"uint"`
			UintPtrFlag     *uint          `flag:"uint_ptr"`
			Uint32Flag      uint32         `flag:"uint32"`
			Uint32PtrFlag   *uint32        `flag:"uint32_ptr"`
			Uint64Flag      uint64         `flag:"uint64"`
			Uint64PtrFlag   *uint64        `flag:"uint64_ptr"`
			Float32Flag     float32        `flag:"float32"`
			Float32PtrFlag  *float32       `flag:"float32_ptr"`
			Float64Flag     float64        `flag:"float64"`
			Float64PtrFlag  *float64       `flag:"float64_ptr"`
			StrFlag         string         `flag:"str"`
			StrPtrFlag      *string        `flag:"str_ptr"`
			DurationFlag    time.Duration  `flag:"duration"`
			DurationPtrFlag *time.Duration `flag:"duration_ptr"`
		}

var tests []testCase = []testCase{
	{
		desc: "all types",
		testStruct: allTypesStruct{
			BoolFlag:        true,
			BoolPtrFlag:     new(bool),
			IntFlag:         1,
			IntPtrFlag:      ptrTo(int(2)).(*int),
			Int32Flag:       3,
			Int32PtrFlag:    ptrTo(int32(4)).(*int32),
			Int64Flag:       5,
			Int64PtrFlag:    ptrTo(int64(6)).(*int64),
			UintFlag:        7,
			UintPtrFlag:     ptrTo(uint(8)).(*uint),
			Uint32Flag:      9,
			Uint32PtrFlag:   ptrTo(uint32(10)).(*uint32),
			Uint64Flag:      11,
			Uint64PtrFlag:   ptrTo(uint64(12)).(*uint64),
			Float32Flag:     13,
			Float32PtrFlag:  ptrTo(float32(14)).(*float32),
			Float64Flag:     15,
			Float64PtrFlag:  ptrTo(float64(16)).(*float64),
			StrFlag:         "init",
			StrPtrFlag:      ptrTo("init_ptr").(*string),
			DurationFlag:    1 * time.Second,
			DurationPtrFlag: ptrTo(2 * time.Second).(*time.Duration),
		},
		wantPreParse: map[string]interface{}{
			"bool":         true,
			"bool_ptr":     false,
			"int":          int(1),
			"int_ptr":      int(2),
			"int32":        int32(3),
			"int32_ptr":    int32(4),
			"int64":        int64(5),
			"int64_ptr":    int64(6),
			"uint":         uint(7),
			"uint_ptr":     uint(8),
			"uint32":       uint32(9),
			"uint32_ptr":   uint32(10),
			"uint64":       uint64(11),
			"uint64_ptr":   uint64(12),
			"float32":      float32(13),
			"float32_ptr":  float32(14),
			"float64":      float64(15),
			"float64_ptr":  float64(16),
			"str":          "init",
			"str_ptr":      "init_ptr",
			"duration":     1 * time.Second,
			"duration_ptr": 2 * time.Second,
		},
		args: []string{
			"--bool=false",
			"--bool_ptr",
			"--int=100",
			"--int_ptr=101",
			"--int32=102",
			"--int32_ptr=103",
			"--int64=104",
			"--int64_ptr=105",
			"--uint=106",
			"--uint_ptr=107",
			"--uint32=108",
			"--uint32_ptr=109",
			"--uint64=110",
			"--uint64_ptr=111",
			"--float32=112",
			"--float32_ptr=113",
			"--float64=114",
			"--float64_ptr=115",
			"--str=parsed",
			"--str_ptr=parsedPtr",
			"--duration=1m",
			"--duration_ptr=1h",
		},
		wantStruct: allTypesStruct{
			BoolFlag:        false,
			BoolPtrFlag:     ptrTo(true).(*bool),
			IntFlag:         100,
			IntPtrFlag:      ptrTo(int(101)).(*int),
			Int32Flag:       102,
			Int32PtrFlag:    ptrTo(int32(103)).(*int32),
			Int64Flag:       104,
			Int64PtrFlag:    ptrTo(int64(105)).(*int64),
			UintFlag:        106,
			UintPtrFlag:     ptrTo(uint(107)).(*uint),
			Uint32Flag:      108,
			Uint32PtrFlag:   ptrTo(uint32(109)).(*uint32),
			Uint64Flag:      110,
			Uint64PtrFlag:   ptrTo(uint64(111)).(*uint64),
			Float32Flag:     112,
			Float32PtrFlag:  ptrTo(float32(113)).(*float32),
			Float64Flag:     114,
			Float64PtrFlag:  ptrTo(float64(115)).(*float64),
			StrFlag:         "parsed",
			StrPtrFlag:      ptrTo("parsedPtr").(*string),
			DurationFlag:    1 * time.Minute,
			DurationPtrFlag: ptrTo(1 * time.Hour).(*time.Duration),
		},
	},
	{
		desc: "fail on uninitialized ptr",
		testStruct: struct {
			Field1 *string `flag:"field1"`
		}{
			Field1: nil,
		},
		wantRegisterErr: errors.New(`unable to register flag for field struct { Field1 *string "flag:\"field1\"" }.Field1: uninitialized pointer`),
	},
	{
		desc: "nested structs",
		testStruct: struct {
			Field1 string `flag:"field1"`
			Nested nested
		}{
			Field1: "field1_value",
			Nested: nested{
				Nested: nested1{
					Nested: nested2{
						Field: "nested_field_value",
					},
				},
			},
		},
		wantPreParse: map[string]interface{}{
			"field1":      "field1_value",
			"nested_flag": "nested_field_value",
		},
		args: []string{
			"--field1=newfield1_value",
			"--nested_flag=newnested_field_value",
		},
		wantStruct: struct {
			Field1 string `flag:"field1"`
			Nested nested
		}{
			Field1: "newfield1_value",
			Nested: nested{
				Nested: nested1{
					Nested: nested2{
						Field: "newnested_field_value",
					},
				},
			},
		},
	},
	{
		desc: "nested struct without public fields",
		testStruct: struct{
			Field1 string `flag:"field1"`
			Nested struct{
				privField string
			}
		}{
			Field1: "foo",
			Nested: struct{
				privField string
			}{
				privField: "foo",
			},
		},
		wantRegisterErr: errors.New(`unable to register flag for field struct { Field1 string "flag:\"field1\""; Nested struct { privField string } }.Nested: unable to register flags for type "struct { privField string }": no exported fields`),
	},
	{
		desc: "overflow int32",
		testStruct: struct{
			I int32 `flag:"i"`
		}{
			I: 1,
		},
		wantPreParse: map[string]interface{}{
			"i": int32(1),
		},
		args: []string{
			"--i=2200000000",
		},
		wantParseErr: errors.New(`invalid value "2200000000" for flag -i: strconv.ParseInt: parsing "2200000000": value out of range`),
	},
	{
		desc: "overflow uint32",
		testStruct: struct{
			I uint32 `flag:"i"`
		}{
			I: 1,
		},
		wantPreParse: map[string]interface{}{
			"i": uint32(1),
		},
		args: []string{
			"--i=4300000000",
		},
		wantParseErr: errors.New(`invalid value "4300000000" for flag -i: strconv.ParseUint: parsing "4300000000": value out of range`),
	},
	{
		desc: "fail non-struct",
		testStruct: "foo",
		wantRegisterErr: errors.New(`unable to register flags for "string": not a struct type`),
	},
	{
		desc: "fail missing flag tag",
		testStruct: struct{
			S string
		}{
			S: "foo",
		},
		wantRegisterErr: errors.New(`unable to register flag for field struct { S string }.S: no "flag" tag found`),
	},
	{
		desc: "fail no exported fields",
		testStruct: struct{
			s string 
		}{
			s: "foo",
		},
		wantRegisterErr: errors.New(`unable to register flags for type "struct { s string }": no exported fields`),
	},
	{
		desc: "test customType",
		testStruct: struct{
			S *customType `flag:"custom"`
		}{
			S: newCustomType("foo"),
		},
		wantPreParse: map[string]interface{}{
			"custom":      newCustomType("foo"),
		},
		args: []string{
			"--custom=bar",
		},
		wantStruct: struct{
			S *customType `flag:"custom"`
		}{
			S: newCustomType("bar"),
		},
		opts: []Option{FlagType(new(customType), newCustomFlag)},
	},
	{
		desc: "test custom tag",
		testStruct: struct{
			S string `flagname:"myflag"`
		}{
			S: "foo",
		},
		wantPreParse: map[string]interface{}{
			"myflag": "foo",
		},
		args: []string{
			"--myflag=bar",
		},
		wantStruct: struct{
			S string `flagname:"myflag"`
		}{
			S: "bar",
		},
		opts: []Option{TagName("flagname")},
	},
	{
		desc: "test custom prefix",
		testStruct: struct{
			S string `flag:"myflag"`
		}{
			S: "foo",
		},
		wantPreParse: map[string]interface{}{
			"lib_myflag": "foo",
		},
		args: []string{
			"--lib_myflag=bar",
		},
		wantStruct: struct{
			S string `flag:"myflag"`
		}{
			S: "bar",
		},
		opts: []Option{FlagPrefix("lib_")},
	},
}

func TestRegisterAndLoadFlags(t *testing.T) {
	for _, tc := range tests {
		if err := runTestCase(tc); err != nil {
			t.Errorf("TestRegisterAndLoadFlags %q failed: %v", tc.desc, err)
		}
	}
}
