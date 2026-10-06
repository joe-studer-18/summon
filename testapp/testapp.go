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
	inj := digo.Injector{}

	fmt.Println(inj.Get[processors.BarName]())

	procs := []Processor{
		inj.Get[processors.FooProcessor](),
	}

	for _, proc := range procs {
		proc.Process()
	}
}
