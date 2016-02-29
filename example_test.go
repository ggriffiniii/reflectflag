package reflectflag_test

import (
	"flag"
	"fmt"
	"github.com/ggriffiniii/reflectflag"
	"time"
)

type MyOptions struct {
	Msg *string `flag:"msg"`
	Elapsed time.Duration `flag:"elapsed"`
	Values []int `flag:"values"`
}

func Example() {
	flags := flag.NewFlagSet("example", flag.ContinueOnError)
	reflectflag.RegisterFlags(flags, MyOptions{})
	flags.Parse([]string{
		"--msg=my message",
		"--elapsed=1h15m",
		"--values=100,15,20",
	})
	var opts MyOptions
	reflectflag.LoadFromFlags(flags, &opts)
	fmt.Printf("Msg: %v\n", *opts.Msg)
	fmt.Printf("Elapsed: %v\n", opts.Elapsed)
	fmt.Printf("Values: %v\n", opts.Values)
	// Output:
	// Msg: my message
	// Elapsed: 1h15m0s
	// Values: [100 15 20]
}
