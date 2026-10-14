package summon

import (
	"errors"
	"fmt"
	"reflect"
)

var bindMap = make(map[reflect.Type]reflect.Type)

var ErrBindCollision = errors.New("bind collision")

func Binds[I any, T any]() error {
	interfaceType := reflect.TypeFor[I]()
	implType := reflect.TypeFor[T]()

	if _, ok := bindMap[interfaceType]; !ok {
		bindMap[interfaceType] = implType
		return nil
	} else {
		panic(fmt.Errorf("%w: %v already bound to %v. Binding to %T results in a collision.", ErrBindCollision, reflect.TypeOf(new(I)).Elem(), bindMap[interfaceType].String(), *new(T)))
	}
}
