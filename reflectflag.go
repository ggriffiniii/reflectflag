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
	hasExportedField := false
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" {
			// skip non-exported fields
			continue
		}
		hasExportedField = true
		if err := registerStructField(flags, sf, v.Field(i), opts); err != nil {
			return fmt.Errorf("unable to register flag for field %s.%s: %v", typ, sf.Name, err)
		}
	}
	if !hasExportedField {
		return fmt.Errorf("unable to register flags for type %q: no exported fields", typ)
	}
	return nil
}

func registerStructField(flags *flag.FlagSet, sf reflect.StructField, v reflect.Value, opts options) error {
	
	tag := sf.Tag.Get(opts.tagName)
	f, ok := opts.ftypes[v.Type()]
	if !ok {
		switch v.Type().Kind() {
		case reflect.Ptr:
			if v.IsNil() {
				return errors.New("uninitialized pointer")
			}
			return registerStructField(flags, sf, v.Elem(), opts)
		case reflect.Struct:
			return registerStructFields(flags, v, opts)
		}
		return errors.New("no flag factory registered")
	}
	if tag == "" {
		return fmt.Errorf("no %q tag found", opts.tagName)
	}
	flags.Var(f(v.Interface()), opts.flagPrefix + tag, fmt.Sprintf("Set %s.%s", v.Type(), sf.Name))
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
	hasExportedField := false
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" {
			// skip non-exported fields
			continue
		}
		hasExportedField = true
		if err := loadFromStructField(flags, sf, v.Field(i), opts); err != nil {
			return fmt.Errorf("unable to load flag for field %s.%s: %v", typ, sf.Name, err)
		}
	}
	if !hasExportedField {
		return fmt.Errorf("unable to load flags for type %q: no exported fields", typ)
	}
	return nil
}

func loadFromStructField(flags *flag.FlagSet, sf reflect.StructField, v reflect.Value, opts options) error {
	_, ok := opts.ftypes[v.Type()]
	if !ok {
		switch v.Type().Kind() {
		case reflect.Ptr:
			next := reflect.New(v.Type().Elem())
			err := loadFromStructField(flags, sf, next.Elem(), opts)
			if err != nil {
				return err
			}
			v.Set(next)
			return nil
		case reflect.Struct:
			return loadFromStructFields(flags, v, opts)
		}
		return errors.New("no flag factory registered")
	}
	tag := sf.Tag.Get(opts.tagName)
	if tag == "" {
		return fmt.Errorf("no %q tag found", opts.tagName)
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
	v.Set(reflect.ValueOf(flagGetter.Get()))
	return nil
}
