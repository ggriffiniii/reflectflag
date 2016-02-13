package reflectflag

import (
	"flag"
	"fmt"
	"reflect"
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

// FlagGetterFactory accepts an interface type and returns a flag.Getter. When
// a FlagGetterFactory is registered with a FlagType Option it will always be
// invoked with an interface that's a pointer to the registered type.
type FlagGetterFactory func(interface{}) flag.Getter

func FlagType(typ interface{}, factory FlagGetterFactory) Option {
	t := reflect.TypeOf(typ)
	if t.Kind() != reflect.Ptr {
		t = reflect.PtrTo(t)
	}
	return flagTypeOpt{
		typ:     t,
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
	ftypes  map[reflect.Type]FlagGetterFactory
}

func RegisterFlags(flags *flag.FlagSet, s interface{}, opts ...Option) error {
	// Initialize the default options combined with the passed in options.
	// Passed in options override defaults.
	o := options{
		ftypes: map[reflect.Type]FlagGetterFactory{},
	}
	default_opts := []Option{TagName("flag"), FlagType("string", newStringValue)}
	for _, x := range append(default_opts, opts...) {
		x.set(&o)
	}

	sv := reflect.ValueOf(s)
	st := sv.Type()
	if sv.Kind() != reflect.Struct {
		return fmt.Errorf("unable to register flags for non-struct type %q", st)
	}
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		if sf.PkgPath != "" {
			// skip non-exported fields
			continue
		}
		tag := sf.Tag.Get(o.tagName)
		if tag == "" {
			return fmt.Errorf("unable to register flag for field %q of type %q; no %q tag specified", sf.Name, st, o.tagName)
		}
		flagFactory, ok := o.ftypes[sf.Type]
		if !ok {
			return fmt.Errorf("unable to register flag for field %q of type %q; no flag factory registered for type: %q", sf.Name, st, sf.Type)
		}
		
	}
	return nil
}
