package reflectflag

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"time"
)

type Option interface {
	set(*options)
}

func TagName(tag string) Option {
	return tagOpt(tag)
}

type tagOpt string

func (o tagOpt) set(opts *options) {
	opts.tagName = string(o)
}

func FlagPrefix(prefix string) Option {
	return flagPrefixOpt(prefix)
}

type flagPrefixOpt string
func (o flagPrefixOpt) set(opts *options) {
	opts.flagPrefix = string(o)
}

// FlagGetterFactory accepts an interface type and returns a flag.Getter. When
// a FlagGetterFactory is registered with a FlagType Option it will always be
// invoked with an interface that matches the registered type.
type FlagGetterFactory func(interface{}) flag.Getter

func FlagType(typ interface{}, factory FlagGetterFactory) Option {
	return flagTypeOpt{
		typ:     reflect.TypeOf(typ),
		factory: factory,
	}
}

type flagTypeOpt struct {
	typ     reflect.Type
	factory FlagGetterFactory
}

func (o flagTypeOpt) set(opts *options) {
	opts.ftypes[o.typ] = o.factory
}

type options struct {
	tagName string
	flagPrefix string
	ftypes  map[reflect.Type]FlagGetterFactory
}

func getOpts(opts ...Option) options {
	o := options{
		ftypes: map[reflect.Type]FlagGetterFactory{},
	}
	default_opts := []Option{
		TagName("flag"),
		FlagType(true, newBoolValue),
		FlagType(int(1), newIntValue),
		FlagType(int32(1), newInt32Value),
		FlagType(int64(1), newInt64Value),
		FlagType(uint(1), newUintValue),
		FlagType(uint32(1), newUint32Value),
		FlagType(uint64(1), newUint64Value),
		FlagType(float32(1), newFloat32Value),
		FlagType(float64(1), newFloat64Value),
		FlagType("string", newStringValue),
		FlagType(time.Second, newDurationValue),
	}
	for _, x := range append(default_opts, opts...) {
		x.set(&o)
	}
	return o
}

func RegisterFlags(flags *flag.FlagSet, s interface{}, opts ...Option) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("unable to register flags for %q: not a struct type", v.Type())
	}
	o := getOpts(opts...)
	return registerStructFields(flags, v, o)
}

func registerStructFields(flags *flag.FlagSet, v reflect.Value, opts options) error {
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" {
			// skip non-exported fields
			continue
		}
		if err := registerStructField(flags, sf, v.Field(i), opts); err != nil {
			return fmt.Errorf("unable to register flag for field %s.%s: %v", typ, sf.Name, err)
		}
	}
	return nil
}

func registerStructField(flags *flag.FlagSet, sf reflect.StructField, v reflect.Value, opts options) error {
	tag := sf.Tag.Get(opts.tagName)
	if tag == "" {
		derefV := derefFully(v)
		if v.Kind() == reflect.Struct {
			return registerStructFields(flags, derefV, opts)
		}
		return nil
	}
	f, fv := factoryForValue(v, opts)
	if f == nil {
		return errors.New("no flag factory registered")
	}
	flags.Var(f(fv.Interface()), opts.flagPrefix + tag, fmt.Sprintf("Set %s.%s", v.Type(), sf.Name))
	return nil
}

func LoadFromFlags(flags *flag.FlagSet, s interface{}, opts ...Option) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("unable to load from flags for %q: not a pointer to a struct", v.Type())
	}
	if !v.Elem().CanSet() {
		return fmt.Errorf("unable to load from flags for %q: struct is not settable", v.Type())
	}
	o := getOpts(opts...)
	return loadFromStructFields(flags, v.Elem(), o)
}

func loadFromStructFields(flags *flag.FlagSet, v reflect.Value, opts options) error {
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" {
			// skip non-exported fields
			continue
		}
		if err := loadFromStructField(flags, sf, v.Field(i), opts); err != nil {
			return fmt.Errorf("unable to load flag for field %s.%s: %v", typ, sf.Name, err)
		}
	}
	return nil
}

func loadFromStructField(flags *flag.FlagSet, sf reflect.StructField, v reflect.Value, opts options) error {
	tag := sf.Tag.Get(opts.tagName)
	if tag == "" {
		derefV := derefFully(v)
		if derefV.Kind() == reflect.Struct {
			return loadFromStructFields(flags, derefV, opts)
		}
		return nil
	}
	f, _ := factoryForValue(v, opts)
	if f == nil {
		return errors.New("no flag factory registered")
	}
	flagName := opts.flagPrefix + tag
	flg := flags.Lookup(flagName)
	if flg == nil {
		return fmt.Errorf("unable to lookup flag %q. Was RegisterFlags called?", flagName)
	}
	flgVal := flg.Value
	flagGetter, ok := flgVal.(flag.Getter)
	if !ok {
		return fmt.Errorf("flag %q doesn't implement the flag.Getter interface", flagName)
	}
	newV, err := convertValueTo(reflect.ValueOf(flagGetter.Get()), v.Type())
	if err != nil {
		return err
	}
	v.Set(newV)
	return nil
}

// derefFully dereferences pointer values until it reaches a non-pointer value.
// If a nil pointer is reached it returns the zero value of the eventual
// non-pointer type. If a non-pointer value is provided it is returned unchanged.
func derefFully(v reflect.Value) reflect.Value {
	for v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			v = reflect.Zero(v.Type().Elem())
		}
		v = v.Elem()
	}
	return v
}

// convertValueTo converts the value to the specified type and returns the new
// value. The value and type must differ only by pointer indirection. If the
// specified type requires additional pointers new pointers will be created
// pointing to the address of v. If fewer pointers are required the value will
// be dereferenced the necessary amount. If a nil pointer is encountered that
// requires dereferencing the types zero value will be returned.
func convertValueTo(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	if v.Type() == typ {
		return v, nil
	}
	typBaseType := typ
	typPtrDepth := 0
	for typBaseType.Kind() == reflect.Ptr {
		typPtrDepth++
		typBaseType = typBaseType.Elem()
	}
	vBaseType := v.Type()
	vPtrDepth := 0
	for vBaseType.Kind() == reflect.Ptr {
		vPtrDepth++
		vBaseType = vBaseType.Elem()
	}
	switch {
	case vBaseType != typBaseType:
		return reflect.Zero(typ), fmt.Errorf("cannot convert between %v and %v: differ by more than pointer indirection", v.Type(), typ)
	case vPtrDepth == typPtrDepth:
		return v, nil
	case vPtrDepth < typPtrDepth:
		for v.Type() != typ {
			x := reflect.New(v.Type())
			x.Elem().Set(v)
			v = x
		}
		return v, nil
	case vPtrDepth > typPtrDepth:
		for v.Type() != typ {
			if v.IsNil() {
				v = reflect.Zero(v.Type().Elem())
			} else {
				v = v.Elem()
			}
		}
		return v, nil
	}
	return reflect.Zero(typ), errors.New("unreachable")
}

// factoryForValue returns the registered flag factory for the specified value
// and the value that should be provided to the factory function on flag
// initialization. If no flag factory is registered for the specified type the
// returned factory will be nil.
func factoryForValue(v reflect.Value, opts options) (FlagGetterFactory, reflect.Value) {
	for typ, factory := range opts.ftypes {
		cv, err := convertValueTo(v, typ)
		if err != nil {
			continue
		}
		return factory, cv
	}
	return nil, v
}
