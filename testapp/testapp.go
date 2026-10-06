package main

import (
	"github.com/digo"
	"github.com/testapp/processors"
)

type Processor interface {
	Process() bool
}

func main() {
	inj := digo.Injector{}

	procs := []Processor{
		inj.Get[processors.FooProcessor](),
	}

	for _, proc := range procs {
		proc.Process()
	}
}
