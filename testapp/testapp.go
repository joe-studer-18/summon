package main

import (
	"fmt"

	"github.com/digo"
	"github.com/testapp/processors"
)

type Processor interface {
	Process() bool
}

func main() {
	// TODO - Add injector object to support singletons
	// inj := digo.New()

	fmt.Printf("Name: %s\n", digo.Get[processors.FooName]())

	procs := []Processor{
		digo.Get[processors.FooProcessor](),
	}

	for _, proc := range procs {
		proc.Process()
	}
}
