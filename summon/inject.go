package summon

import (
	"errors"
	"fmt"
	"reflect"
)

const provideMethodName = "DiGoProvide"
const invalidReturnTypeErr = "%s.%s does not return a type of %s"

var ErrInvalidReturnType = errors.New("invalid return type")

type Injector struct{}

func (inj *Injector) Get[T any]() (T, error) {
	got, err := inj.get(reflect.TypeFor[T]())
	if err != nil {
		errString := fmt.Sprintf(invalidReturnTypeErr, reflect.TypeOf(*new(T)), provideMethodName, reflect.TypeOf(*new(T)))
		var zero T
		return zero, fmt.Errorf("%w: %s", ErrInvalidReturnType, errString)
	}
	returnable, ok := toValue[T](got)
	if !ok {
		errString := fmt.Sprintf(invalidReturnTypeErr, reflect.TypeOf(*new(T)), provideMethodName, reflect.TypeOf(*new(T)))
		var zero T
		return zero, fmt.Errorf("%w: %s", ErrInvalidReturnType, errString)
	}
	return returnable, nil
}

func (inj *Injector) get(t reflect.Type) (reflect.Value, error) {
	provideable := reflect.New(t)
	providerMethod := provideable.MethodByName(provideMethodName)
	if !providerMethod.IsValid() {
		errString := fmt.Sprintf("%s not implemented for type T <%s>", provideMethodName, provideable.String())
		return reflect.Value{}, errors.New(errString)
	}

	numArgs := providerMethod.Type().NumIn()
	args := make([]reflect.Value, numArgs)
	if numArgs == 0 {
		rawReturnable := providerMethod.Call(args)
		return rawReturnable[0], nil
	}
	for i := range numArgs {
		argType := providerMethod.Type().In(i)
		var err error
		if a, ok := bindMap[argType]; ok {
			argType = a
		}
		args[i], err = inj.get(argType)
		if err != nil {
			return reflect.Value{}, err
		}
	}
	return providerMethod.Call(args)[0], nil
}

func toValue[T any](v reflect.Value) (T, bool) {
	returnable, ok := v.Convert(reflect.TypeFor[T]()).Interface().(T)
	return returnable, ok
}
