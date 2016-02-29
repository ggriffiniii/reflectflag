package reflectflag_test

import (
	"flag"
	"fmt"
	"github.com/ggriffiniii/reflectflag"
)

type MyChoice int
const (
	ChoiceA MyChoice = iota
	ChoiceB
	ChoiceC
)

func (c *MyChoice) Set(s string) error {
	switch s {
	case "ChoiceA":
		*c = ChoiceA
		return nil
	case "ChoiceB":
		*c = ChoiceB
		return nil
	case "ChoiceC":
		*c = ChoiceC
		return nil
	}
	return fmt.Errorf("invalid choice: %q", s)
}

func (c *MyChoice) String() string {
	switch *c {
	case ChoiceA: return "ChoiceA"
	case ChoiceB: return "ChoiceB"
	case ChoiceC: return "ChoiceC"
	}
	return "invalid"
}

func (c *MyChoice) Get() interface{} {
	x := *c
	return &x
}

func choiceFactory(c interface{}) flag.Getter {
	r := c.(MyChoice)
	return &r
}

type MyOptions2 struct {
	First MyChoice `flag:"first"`
	Second *MyChoice `flag:"second"`
	Extra []MyChoice `flag:"extra"`
	ExtraExtra []*MyChoice `flag:"extraextra"`
}

// ExampleCustomType demonstrates how to use reflectflag with a custom type.
func Example_customType() {
	flags := flag.NewFlagSet("example", flag.ContinueOnError)
	reflectOpts := []reflectflag.Option{
		reflectflag.FlagType(MyChoice(0), choiceFactory),
	}
	reflectflag.RegisterFlags(flags, MyOptions2{}, reflectOpts...)
	flags.Parse([]string{
		"--first=ChoiceA",
		"--second=ChoiceC",
		"--extra=ChoiceB,ChoiceA",
		"--extraextra=ChoiceC,ChoiceB,ChoiceA",
	})
	var opts MyOptions2
	reflectflag.LoadFromFlags(flags, &opts, reflectOpts...)
	fmt.Printf("First: %v\n", &opts.First)
	fmt.Printf("Second: %v\n", opts.Second)
	for i, v := range opts.Extra {
		fmt.Printf("Extra[%d]: %v\n", i, &v)
	}
	for i, v := range opts.ExtraExtra {
		fmt.Printf("ExtraExtra[%d]: %v\n", i, v)
	}
	// Output:
	// First: ChoiceA
	// Second: ChoiceC
	// Extra[0]: ChoiceB
	// Extra[1]: ChoiceA
	// ExtraExtra[0]: ChoiceC
	// ExtraExtra[1]: ChoiceB
	// ExtraExtra[2]: ChoiceA
}
