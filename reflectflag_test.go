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
				continue // skip unexported fields
			}
			if !deepEqual(va.Field(i).Interface(), vb.Field(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if va.Len() != vb.Len() {
			return false
		}
		for i := 0; i < va.Len(); i++ {
			if !deepEqual(va.Index(i).Interface(), vb.Index(i).Interface()) {
				fmt.Printf("Index %v not equal\na: %v\nb: %v\n", i, va.Index(i).Interface(), vb.Index(i).Interface())
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
	case "foo":
		x = customType(1)
	case "bar":
		x = customType(2)
	case "baz":
		x = customType(3)
	}
	return &x
}

func (ct *customType) GoString() string {
	return fmt.Sprintf("%v", int32(*ct))
}

type customFlag struct {
	t *customType
}

func newCustomFlag(c interface{}) flag.Getter {
	return &customFlag{t: c.(*customType)}
}

func (c *customFlag) String() string {
	switch int32(*c.t) {
	case 1:
		return "foo"
	case 2:
		return "bar"
	case 3:
		return "baz"
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
	fmt.Println("Set custom type to: %v", *c.t)
	return errors.New("unable to set customType")
}

func (c *customFlag) Get() interface{} {
	x := *c.t
	return &x
}

type testCase struct {
	desc            string
	testStruct      interface{}
	wantRegisterErr error
	wantPreParse    map[string]interface{}
	args            []string
	wantParseErr    error
	wantStruct      interface{}
	wantLoadErr     error
	opts            []Option
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
	BoolFlag             bool             `flag:"bool"`
	BoolSliceFlag        []bool           `flag:"bool_slice"`
	BoolPtrFlag          *bool            `flag:"bool_ptr"`
	BoolSlicePtrFlag     []*bool          `flag:"bool_slice_ptr"`
	IntFlag              int              `flag:"int"`
	IntSliceFlag         []int            `flag:"int_slice"`
	IntPtrFlag           *int             `flag:"int_ptr"`
	IntSlicePtrFlag      []*int           `flag:"int_slice_ptr"`
	Int32Flag            int32            `flag:"int32"`
	Int32SliceFlag       []int32          `flag:"int32_slice"`
	Int32PtrFlag         *int32           `flag:"int32_ptr"`
	Int32SlicePtrFlag    []*int32         `flag:"int32_slice_ptr"`
	Int64Flag            int64            `flag:"int64"`
	Int64SliceFlag       []int64          `flag:"int64_slice"`
	Int64PtrFlag         *int64           `flag:"int64_ptr"`
	Int64SlicePtrFlag    []*int64         `flag:"int64_slice_ptr"`
	UintFlag             uint             `flag:"uint"`
	UintSliceFlag        []uint           `flag:"uint_slice"`
	UintPtrFlag          *uint            `flag:"uint_ptr"`
	UintSlicePtrFlag     []*uint          `flag:"uint_slice_ptr"`
	Uint32Flag           uint32           `flag:"uint32"`
	Uint32SliceFlag      []uint32         `flag:"uint32_slice"`
	Uint32PtrFlag        *uint32          `flag:"uint32_ptr"`
	Uint32SlicePtrFlag   []*uint32        `flag:"uint32_slice_ptr"`
	Uint64Flag           uint64           `flag:"uint64"`
	Uint64SliceFlag      []uint64         `flag:"uint64_slice"`
	Uint64PtrFlag        *uint64          `flag:"uint64_ptr"`
	Uint64SlicePtrFlag   []*uint64        `flag:"uint64_slice_ptr"`
	Float32Flag          float32          `flag:"float32"`
	Float32SliceFlag     []float32        `flag:"float32_slice"`
	Float32PtrFlag       *float32         `flag:"float32_ptr"`
	Float32SlicePtrFlag  []*float32       `flag:"float32_slice_ptr"`
	Float64Flag          float64          `flag:"float64"`
	Float64SliceFlag     []float64        `flag:"float64_slice"`
	Float64PtrFlag       *float64         `flag:"float64_ptr"`
	Float64SlicePtrFlag  []*float64       `flag:"float64_slice_ptr"`
	StrFlag              string           `flag:"str"`
	StrSliceFlag         []string         `flag:"str_slice"`
	StrPtrFlag           *string          `flag:"str_ptr"`
	StrSlicePtrFlag      []*string        `flag:"str_slice_ptr"`
	DurationFlag         time.Duration    `flag:"duration"`
	DurationSliceFlag    []time.Duration  `flag:"duration_slice"`
	DurationPtrFlag      *time.Duration   `flag:"duration_ptr"`
	DurationSlicePtrFlag []*time.Duration `flag:"duration_slice_ptr"`
}

var tests []testCase = []testCase{
	{
		desc: "all types",
		testStruct: allTypesStruct{
			BoolFlag:            true,
			BoolSliceFlag:       []bool{true, false, true},
			BoolPtrFlag:         new(bool),
			BoolSlicePtrFlag:    []*bool{ptrTo(false).(*bool), ptrTo(true).(*bool), ptrTo(true).(*bool)},
			IntFlag:             1,
			IntSliceFlag:        []int{1, 2, 3},
			IntPtrFlag:          ptrTo(int(2)).(*int),
			IntSlicePtrFlag:     []*int{ptrTo(int(2)).(*int), ptrTo(int(4)).(*int)},
			Int32Flag:           3,
			Int32SliceFlag:      []int32{10, 20, 30, 40},
			Int32PtrFlag:        ptrTo(int32(4)).(*int32),
			Int32SlicePtrFlag:   []*int32{ptrTo(int32(4)).(*int32), ptrTo(int32(8)).(*int32)},
			Int64Flag:           5,
			Int64SliceFlag:      []int64{5, 4, 3, 2, 1},
			Int64PtrFlag:        ptrTo(int64(6)).(*int64),
			Int64SlicePtrFlag:   []*int64{ptrTo(int64(6)).(*int64), ptrTo(int64(3)).(*int64)},
			UintFlag:            7,
			UintSliceFlag:       []uint{7, 8, 9},
			UintPtrFlag:         ptrTo(uint(8)).(*uint),
			UintSlicePtrFlag:    []*uint{ptrTo(uint(8)).(*uint)},
			Uint32Flag:          9,
			Uint32SliceFlag:     []uint32{9, 18, 27},
			Uint32PtrFlag:       ptrTo(uint32(10)).(*uint32),
			Uint32SlicePtrFlag:  []*uint32{ptrTo(uint32(10)).(*uint32), ptrTo(uint32(20)).(*uint32)},
			Uint64Flag:          11,
			Uint64SliceFlag:     []uint64{99, 98, 97},
			Uint64PtrFlag:       ptrTo(uint64(12)).(*uint64),
			Uint64SlicePtrFlag:  []*uint64{ptrTo(uint64(89)).(*uint64), ptrTo(uint64(88)).(*uint64)},
			Float32Flag:         13,
			Float32SliceFlag:    []float32{50, 25, 75},
			Float32PtrFlag:      ptrTo(float32(14)).(*float32),
			Float32SlicePtrFlag: []*float32{ptrTo(float32(14)).(*float32), ptrTo(float32(28)).(*float32)},
			Float64Flag:         15,
			Float64SliceFlag:    []float64{1, 3, 7, 9},
			Float64PtrFlag:      ptrTo(float64(16)).(*float64),
			Float64SlicePtrFlag: []*float64{ptrTo(float64(13)).(*float64), ptrTo(float64(17)).(*float64)},
			StrFlag:             "init",
			StrSliceFlag:        []string{"init", "parse", "exec"},
			StrPtrFlag:          ptrTo("init_ptr").(*string),
			StrSlicePtrFlag: []*string{ptrTo("init_ptr").(*string),
				ptrTo("parse_ptr").(*string),
				ptrTo("exec_ptr").(*string)},
			DurationFlag:      1 * time.Second,
			DurationSliceFlag: []time.Duration{time.Second, time.Minute, time.Hour},
			DurationPtrFlag:   ptrTo(2 * time.Second).(*time.Duration),
			DurationSlicePtrFlag: []*time.Duration{ptrTo(3 * time.Second).(*time.Duration),
				ptrTo(5 * time.Second).(*time.Duration)},
		},
		wantPreParse: map[string]interface{}{
			"bool":               true,
			"bool_slice":         []interface{}{true, false, true},
			"bool_ptr":           false,
			"bool_slice_ptr":     []interface{}{false, true, true},
			"int":                int(1),
			"int_slice":          []interface{}{int(1), int(2), int(3)},
			"int_ptr":            int(2),
			"int_slice_ptr":      []interface{}{int(2), int(4)},
			"int32":              int32(3),
			"int32_slice":        []interface{}{int32(10), int32(20), int32(30), int32(40)},
			"int32_ptr":          int32(4),
			"int32_slice_ptr":    []interface{}{int32(4), int32(8)},
			"int64":              int64(5),
			"int64_slice":        []interface{}{int64(5), int64(4), int64(3), int64(2), int64(1)},
			"int64_ptr":          int64(6),
			"int64_slice_ptr":    []interface{}{int64(6), int64(3)},
			"uint":               uint(7),
			"uint_slice":         []interface{}{uint(7), uint(8), uint(9)},
			"uint_ptr":           uint(8),
			"uint_slice_ptr":     []interface{}{uint(8)},
			"uint32":             uint32(9),
			"uint32_slice":       []interface{}{uint32(9), uint32(18), uint32(27)},
			"uint32_ptr":         uint32(10),
			"uint32_slice_ptr":   []interface{}{uint32(10), uint32(20)},
			"uint64":             uint64(11),
			"uint64_slice":       []interface{}{uint64(99), uint64(98), uint64(97)},
			"uint64_ptr":         uint64(12),
			"uint64_slice_ptr":   []interface{}{uint64(89), uint64(88)},
			"float32":            float32(13),
			"float32_slice":      []interface{}{float32(50), float32(25), float32(75)},
			"float32_ptr":        float32(14),
			"float32_slice_ptr":  []interface{}{float32(14), float32(28)},
			"float64":            float64(15),
			"float64_slice":      []interface{}{float64(1), float64(3), float64(7), float64(9)},
			"float64_ptr":        float64(16),
			"float64_slice_ptr":  []interface{}{float64(13), float64(17)},
			"str":                "init",
			"str_slice":          []interface{}{"init", "parse", "exec"},
			"str_ptr":            "init_ptr",
			"str_slice_ptr":      []interface{}{"init_ptr", "parse_ptr", "exec_ptr"},
			"duration":           1 * time.Second,
			"duration_slice":     []interface{}{time.Second, time.Minute, time.Hour},
			"duration_ptr":       2 * time.Second,
			"duration_slice_ptr": []interface{}{3 * time.Second, 5 * time.Second},
		},
		args: []string{
			"--bool=false",
			"--bool_slice=false,false,true",
			"--bool_ptr",
			"--bool_slice_ptr=false,true,false",
			"--int=100",
			"--int_slice=3,2,1",
			"--int_ptr=101",
			"--int_slice_ptr=4,8",
			"--int32=102",
			"--int32_slice=5,10,15,20",
			"--int32_ptr=103",
			"--int32_slice_ptr=32",
			"--int64=104",
			"--int64_slice=",
			"--int64_ptr=105",
			"--int64_slice_ptr=3,6",
			"--uint=106",
			"--uint_slice=14,16,18",
			"--uint_ptr=107",
			"--uint_slice_ptr=108",
			"--uint32=108",
			"--uint32_slice=51,52",
			"--uint32_ptr=109",
			"--uint32_slice_ptr=99,88",
			"--uint64=110",
			"--uint64_slice=5,10,5,10,5",
			"--uint64_ptr=111",
			"--uint64_slice_ptr=43,11",
			"--float32=112",
			"--float32_slice=25.1,90.2",
			"--float32_ptr=113",
			"--float32_slice_ptr=35.9,0",
			"--float64=114",
			"--float64_slice=1.1,3.2,7.3,9.4",
			"--float64_ptr=115",
			"--float64_slice_ptr=9.2,8.3",
			"--str=parsed",
			"--str_slice=finished,success",
			"--str_ptr=parsedPtr",
			"--str_slice_ptr=run1,run2,done",
			"--duration=1m",
			"--duration_slice=1s,2m,3h",
			"--duration_ptr=1h",
			"--duration_slice_ptr=3h,7h",
		},
		wantStruct: allTypesStruct{
			BoolFlag:            false,
			BoolSliceFlag:       []bool{false, false, true},
			BoolPtrFlag:         ptrTo(true).(*bool),
			BoolSlicePtrFlag:    []*bool{ptrTo(false).(*bool), ptrTo(true).(*bool), ptrTo(false).(*bool)},
			IntFlag:             100,
			IntSliceFlag:        []int{3, 2, 1},
			IntPtrFlag:          ptrTo(int(101)).(*int),
			IntSlicePtrFlag:     []*int{ptrTo(int(4)).(*int), ptrTo(int(8)).(*int)},
			Int32Flag:           102,
			Int32SliceFlag:      []int32{5, 10, 15, 20},
			Int32PtrFlag:        ptrTo(int32(103)).(*int32),
			Int32SlicePtrFlag:   []*int32{ptrTo(int32(32)).(*int32)},
			Int64Flag:           104,
			Int64SliceFlag:      []int64{},
			Int64PtrFlag:        ptrTo(int64(105)).(*int64),
			Int64SlicePtrFlag:   []*int64{ptrTo(int64(3)).(*int64), ptrTo(int64(6)).(*int64)},
			UintFlag:            106,
			UintSliceFlag:       []uint{14, 16, 18},
			UintPtrFlag:         ptrTo(uint(107)).(*uint),
			UintSlicePtrFlag:    []*uint{ptrTo(uint(108)).(*uint)},
			Uint32Flag:          108,
			Uint32SliceFlag:     []uint32{51, 52},
			Uint32PtrFlag:       ptrTo(uint32(109)).(*uint32),
			Uint32SlicePtrFlag:  []*uint32{ptrTo(uint32(99)).(*uint32), ptrTo(uint32(88)).(*uint32)},
			Uint64Flag:          110,
			Uint64SliceFlag:     []uint64{5, 10, 5, 10, 5},
			Uint64PtrFlag:       ptrTo(uint64(111)).(*uint64),
			Uint64SlicePtrFlag:  []*uint64{ptrTo(uint64(43)).(*uint64), ptrTo(uint64(11)).(*uint64)},
			Float32Flag:         112,
			Float32SliceFlag:    []float32{25.1, 90.2},
			Float32PtrFlag:      ptrTo(float32(113)).(*float32),
			Float32SlicePtrFlag: []*float32{ptrTo(float32(35.9)).(*float32), ptrTo(float32(0)).(*float32)},
			Float64Flag:         114,
			Float64SliceFlag:    []float64{1.1, 3.2, 7.3, 9.4},
			Float64PtrFlag:      ptrTo(float64(115)).(*float64),
			Float64SlicePtrFlag: []*float64{ptrTo(float64(9.2)).(*float64), ptrTo(float64(8.3)).(*float64)},
			StrFlag:             "parsed",
			StrSliceFlag:        []string{"finished", "success"},
			StrPtrFlag:          ptrTo("parsedPtr").(*string),
			StrSlicePtrFlag: []*string{ptrTo("run1").(*string),
				ptrTo("run2").(*string),
				ptrTo("done").(*string)},
			DurationFlag:      1 * time.Minute,
			DurationSliceFlag: []time.Duration{time.Second, 2 * time.Minute, 3 * time.Hour},
			DurationPtrFlag:   ptrTo(1 * time.Hour).(*time.Duration),
			DurationSlicePtrFlag: []*time.Duration{ptrTo(3 * time.Hour).(*time.Duration),
				ptrTo(7 * time.Hour).(*time.Duration)},
		},
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
		desc: "overflow int32",
		testStruct: struct {
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
		testStruct: struct {
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
		desc:            "fail non-struct",
		testStruct:      "foo",
		wantRegisterErr: errors.New(`unable to register flags for "string": not a struct type`),
	},
	{
		desc: "fail unknown type",
		testStruct: struct {
			S *customType `flag:"custom"`
		}{
			S: newCustomType("foo"),
		},
		wantPreParse: map[string]interface{}{
			"custom": newCustomType("foo"),
		},
		args: []string{
			"--custom=bar",
		},
		wantStruct: struct {
			S *customType `flag:"custom"`
		}{
			S: newCustomType("bar"),
		},
		wantRegisterErr: errors.New(`unable to register flag for field struct { S *reflectflag.customType "flag:\"custom\"" }.S: no flag factory registered for *reflectflag.customType`),
	},
	{
		desc: "test customType",
		testStruct: struct {
			Base  customType   `flag:"base"`
			Ptr   *customType  `flag:"ptr"`
			Slice []customType `flag:"slice"`
			//SlicePtr []*customType `flag:"slice_ptr"`
		}{
			Base:  *newCustomType("bar"),
			Ptr:   newCustomType("foo"),
			Slice: []customType{*newCustomType("foo"), *newCustomType("baz")},
			//SlicePtr: []*customType{newCustomType("bar"),newCustomType("foo")},
		},
		wantPreParse: map[string]interface{}{
			"base":  newCustomType("bar"),
			"ptr":   newCustomType("foo"),
			"slice": []interface{}{newCustomType("foo"), newCustomType("baz")},
			//"slice_ptr": []interface{}{newCustomType("bar"),newCustomType("foo")},
		},
		args: []string{
			"--base=foo",
			"--ptr=baz",
			"--slice=baz,foo,bar",
			//"--slice_ptr=foo,bar",
		},
		wantStruct: struct {
			Base  customType   `flag:"base"`
			Ptr   *customType  `flag:"ptr"`
			Slice []customType `flag:"slice"`
			//SlicePtr []*customType `flag:"slice_ptr"`
		}{
			Base:  *newCustomType("foo"),
			Ptr:   newCustomType("baz"),
			Slice: []customType{*newCustomType("baz"), *newCustomType("foo"), *newCustomType("bar")},
			//SlicePtr: []*customType{newCustomType("foo"),newCustomType("bar")},
		},
		opts: []Option{FlagType(new(customType), newCustomFlag)},
	},
	{
		desc: "test custom tag",
		testStruct: struct {
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
		wantStruct: struct {
			S string `flagname:"myflag"`
		}{
			S: "bar",
		},
		opts: []Option{TagName("flagname")},
	},
	{
		desc: "test custom prefix",
		testStruct: struct {
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
		wantStruct: struct {
			S string `flag:"myflag"`
		}{
			S: "bar",
		},
		opts: []Option{FlagPrefix("lib_")},
	},
	{
		desc: "test slice quoting",
		testStruct: struct {
			A []string `flag:"a"`
		}{
			A: []string{`with space`, `with,comma`, `with"quote`},
		},
		wantPreParse: map[string]interface{}{
			"a": []interface{}{`with space`, `with,comma`, `with"quote`},
		},
		args: []string{
			`--a="with,comma","with""quote",with space`,
		},
		wantStruct: struct {
			A []string `flag:"a"`
		}{
			A: []string{`with,comma`, `with"quote`, `with space`},
		},
	},
}

func TestRegisterAndLoadFlags(t *testing.T) {
	for _, tc := range tests {
		if err := runTestCase(tc); err != nil {
			t.Errorf("TestRegisterAndLoadFlags %q failed: %v", tc.desc, err)
		}
	}
}

func TestDerefFully(t *testing.T) {
	for _, tc := range []struct {
		in   interface{}
		want interface{}
	}{
		{
			in:   "foo",
			want: "foo",
		},
		{
			in:   ptrTo("foo"),
			want: "foo",
		},
		{
			in:   ptrTo(ptrTo("foo")),
			want: "foo",
		},
	} {
		got := derefFully(reflect.ValueOf(tc.in)).Interface()
		if !deepEqual(got, tc.want) {
			t.Errorf("unexpected output from derefFully: got %v want %v", got, tc.want)
		}
	}
}

func TestConvertValueTo(t *testing.T) {
	for _, tc := range []struct {
		in      interface{}
		want    interface{}
		wantErr error
	}{
		{
			in:   "foo",
			want: "foo",
		},
		{
			in:   ptrTo("foo"),
			want: "foo",
		},
		{
			in:   ptrTo(ptrTo("foo")),
			want: "foo",
		},
		{
			in:   "foo",
			want: ptrTo(ptrTo("foo")),
		},
		{
			in:   ptrTo("foo"),
			want: ptrTo(ptrTo("foo")),
		},
		{
			in:   (**string)(nil),
			want: "",
		},
		{
			in:      ptrTo("foo"),
			want:    ptrTo(ptrTo(int(3))),
			wantErr: errors.New(`cannot convert between *string and **int: differ by more than pointer indirection`),
		},
	} {
		v, err := convertValueTo(reflect.ValueOf(tc.in), reflect.TypeOf(tc.want))
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tc.wantErr) {
			t.Errorf("unexpected error from convertValueTo: got %v want %v", err, tc.wantErr)
		}
		if err != nil {
			continue
		}
		got := v.Interface()
		if !deepEqual(got, tc.want) {
			t.Errorf("unexpected output from convertValueTo: got %v want %v", got, tc.want)
		}
	}
}
