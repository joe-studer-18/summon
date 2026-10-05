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
	return FooProcessor{name.Name()}
}

type FooName struct {
	name string
}

func (f FooName) DiGoProvide() FooName {
	return FooName{name: "DiGo"}
}

func (f FooName) Name() string {
	return f.name
}
