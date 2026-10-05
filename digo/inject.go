package digo

import (
	"fmt"
	"reflect"
)

const provideMethodName = "DiGoProvide"

func Get[T any]() T {
	returnable, ok := toValue[T](get(reflect.TypeFor[T]()))
	if !ok {
		errString := fmt.Sprintf("%s.%s does not return a type of %s", *new(T), provideMethodName, *new(T))
		panic(errString)
	}
	return returnable
}

// Worth noting that this works for hard types - but if I instead take
// an interface as a param I wouldn't be able to realize it
func get(t reflect.Type) reflect.Value {
	provideable := reflect.New(t)
	providerMethod := provideable.MethodByName(provideMethodName)
	if !providerMethod.IsValid() {
		errString := fmt.Sprintf("%s not implemented for type T <%s>", provideMethodName, provideable.String())
		panic(errString)
	}

	numArgs := providerMethod.Type().NumIn()
	args := make([]reflect.Value, numArgs)
	if numArgs == 0 {
		rawReturnable := providerMethod.Call(args)
		return rawReturnable[0]
	}
	for i := range numArgs {
		argType := providerMethod.Type().In(i)
		args[i] = get(argType)
	}
	return providerMethod.Call(args)[0]
}

func toValue[T any](v reflect.Value) (T, bool) {
	returnable, ok := v.Convert(reflect.TypeFor[T]()).Interface().(T)
	return returnable, ok
}
