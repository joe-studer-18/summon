package summon

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	summonMethodName          = "Summon"
	summonSingletonMethodName = "SummonSingleton"
)
const (
	invalidReturnTypeErr = "%s.%s does not return a type of %s"
)

var ErrInvalidSummonerMethod = errors.New("invalid summon method")
var ErrNoSummoner = errors.New("no summoner")
var ErrCircularDependency = errors.New("circular dependency")

type Circle struct {
	singletons map[reflect.Type]reflect.Value
	mu         sync.Mutex
	fetching   *orderedmap.OrderedMap[reflect.Type, struct{}]
}

func NewCircle() Circle {
	return Circle{singletons: make(map[reflect.Type]reflect.Value)}
}

func (circle *Circle) Summon[T any]() (T, error) {
	circle.mu.Lock()
	defer circle.mu.Unlock()
	circle.fetching = orderedmap.New[reflect.Type, struct{}]()

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
	if _, ok := circle.fetching.Get(t); ok {
		var errString strings.Builder
		startPair := circle.fetching.GetPair(t)
		for pair := startPair; pair != nil; pair = pair.Next() {
			k := pair.Key
			fmt.Fprintf(&errString, "%s -> ", k.String())
		}
		errString.WriteString(t.String())
		err := fmt.Errorf("%w: %s", ErrCircularDependency, errString.String())
		return reflect.Value{}, err
	}
	circle.fetching.Set(t, struct{}{})
	if _, ok := bindMap[t]; ok {
		t = bindMap[t]
	}
	if s, ok := circle.singletons[t]; ok {
		circle.fetching.Delete(t)
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
		circle.fetching.Delete(t)
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
	circle.fetching.Delete(t)
	return returnable, nil
}

func getSummonerMethod(summonable reflect.Value) (reflect.Value, bool, error) {
	summonerMethod := summonable.MethodByName(summonMethodName)
	singletonSummonerMethod := summonable.MethodByName(summonSingletonMethodName)
	if summonerMethod.IsValid() && singletonSummonerMethod.IsValid() {
		err := fmt.Errorf("%w: %s and %s implemented for type T %s. Implement at most one summoner method", ErrInvalidSummonerMethod, summonMethodName, summonSingletonMethodName, summonable.String())
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
	return returnType.Kind() == reflect.Pointer
}

func toValue[T any](v reflect.Value) (T, bool) {
	returnable, ok := v.Interface().(T)
	return returnable, ok
}
