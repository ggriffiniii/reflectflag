package reflectflag

import (
	"flag"
	"fmt"
)

type stringValue string

func newStringValue(s interface{}) flag.Getter {
	rs := s.(*string)
	return (*stringValue)(rs)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return fmt.Sprintf("%s", *s) }
