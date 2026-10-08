package summon

import "reflect"

var bindMap = make(map[reflect.Type]reflect.Type)

func Binds[I any, T any]() bool {
	interfaceType := reflect.TypeFor[I]()
	implType := reflect.TypeFor[T]()

	bindMap[interfaceType] = implType
	return true
}
