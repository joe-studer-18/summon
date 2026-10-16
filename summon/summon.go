package summon

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

const (
	summonMethodName          = "Summon"
	summonSingletonMethodName = "SummonSingleton"
)
const invalidReturnTypeErr = "%s.%s does not return a type of %s"

var ErrInvalidSummonerMethod = errors.New("invalid summon method")
var ErrNoSummoner = errors.New("no summoner")

type Circle struct {
	singletons map[reflect.Type]reflect.Value
	mu         sync.Mutex
}

func NewCircle() Circle {
	return Circle{singletons: make(map[reflect.Type]reflect.Value)}
}

func (circle *Circle) Summon[T any]() (T, error) {
	circle.mu.Lock()
	defer circle.mu.Unlock()

	summoned, err := circle.summon(reflect.TypeFor[T]())
	if err != nil {
		var zero T
		return zero, err
	}
	returnable, ok := toValue[T](summoned)
	if !ok {
		errString := fmt.Sprintf(invalidReturnTypeErr, reflect.TypeOf(*new(T)), summonMethodName, reflect.TypeOf(*new(T)))
		var zero T
		return zero, fmt.Errorf("%w: %s", ErrInvalidSummonerMethod, errString)
	}
	return returnable, nil
}

func (circle *Circle) summon(t reflect.Type) (reflect.Value, error) {
	if _, ok := bindMap[t]; ok {
		t = bindMap[t]
	}
	if s, ok := circle.singletons[t]; ok {
		return s, nil
	}
	summonable := reflect.New(t)
	if t.Kind() == reflect.Pointer {
		// Strip the outer pointer off before making the summonable
		summonable = reflect.New(t.Elem())
	}
	summonerMethod, isSingleton, err := getSummonerMethod(summonable)
	if err != nil {
		return reflect.Value{}, err
	}

	numArgs := summonerMethod.Type().NumIn()
	args := make([]reflect.Value, numArgs)
	if numArgs == 0 {
		returnable := summonerMethod.Call(args)[0]
		if isSingleton {
			circle.singletons[t] = returnable
		}
		return returnable, nil
	}
	for i := range numArgs {
		argType := summonerMethod.Type().In(i)
		var err error
		if a, ok := bindMap[argType]; ok {
			argType = a
		}
		args[i], err = circle.summon(argType)

		if err != nil {
			return reflect.Value{}, err
		}
	}
	returnable := summonerMethod.Call(args)[0]
	if isSingleton {
		circle.singletons[t] = returnable
	}
	return returnable, nil
}

func getSummonerMethod(summonable reflect.Value) (reflect.Value, bool, error) {
	summonerMethod := summonable.MethodByName(summonMethodName)
	singletonSummonerMethod := summonable.MethodByName(summonSingletonMethodName)
	if summonerMethod.IsValid() && singletonSummonerMethod.IsValid() {
		err := fmt.Errorf("%w: %s and %s implemented for type T %s. Implement at most one summoner method.", ErrInvalidSummonerMethod, summonMethodName, summonSingletonMethodName, summonable.String())
		return reflect.Value{}, false, err
	}
	if !summonerMethod.IsValid() && !singletonSummonerMethod.IsValid() {
		err := fmt.Errorf("%w: no summoner method for type T %s. Implement one of (%s, %s)", ErrNoSummoner, summonable.String(), summonMethodName, summonSingletonMethodName)
		return reflect.Value{}, false, err
	}
	// exactly one summoner is implemented
	if summonerMethod.IsValid() {
		return summonerMethod, false, nil
	} else {
		if !validateSingletonSummonerMethod(singletonSummonerMethod) {
			err := fmt.Errorf("%w: %s.%s return type of %s is not a pointer", ErrInvalidSummonerMethod, summonable.String(), singletonSummonerMethod, singletonSummonerMethod.Type().Out(0))
			return reflect.Value{}, false, err
		}
		return singletonSummonerMethod, true, nil
	}
}

func validateSingletonSummonerMethod(method reflect.Value) bool {
	returnType := method.Type().Out(0)
	if returnType.Kind() != reflect.Pointer {
		return false
	}
	return true
}

func toValue[T any](v reflect.Value) (T, bool) {
	returnable, ok := v.Interface().(T)
	return returnable, ok
}
