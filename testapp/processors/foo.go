package processors

import "fmt"

type FooProcessor struct {
	name string
}

func (f FooProcessor) Process() bool {
	fmt.Printf("Foo: %s\n", f.name)
	return true
}

func (f FooProcessor) DiGoProvide(name FooName) FooProcessor {
	return FooProcessor{string(name)}
}

type FooName string

func (f FooName) DiGoProvide() FooName {
	return "DiGo"
}

type BarName string

func (f BarName) DiGoProvide() BarName {
	return "DiGoGo"
}
