//go:build unit

package summon

import (
	"crypto/rand"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/joe-studer-18/summon/summon/threadtest"
)

type AnInterface interface {
	Do() string
}

type AnImplementation struct{}

func (a AnImplementation) Do() string {
	return "AnImplementation"
}

func (a AnImplementation) Summon() AnImplementation {
	return AnImplementation{}
}

type RootNode struct {
	name string
}

func (f RootNode) Process() bool {
	fmt.Printf("Foo: %s\n", f.name)
	return true
}

func (f RootNode) Summon(name LeafNode, i AnInterface) RootNode {
	return RootNode{fmt.Sprintf("%s.%s", string(name), i.Do())}
}

type LeafNode string

func (f LeafNode) Summon() LeafNode {
	return "DiGo"
}

type Pointer struct{}

type TwoSummonMethods struct{}

func (t TwoSummonMethods) Summon() TwoSummonMethods {
	return TwoSummonMethods{}
}

func (t *TwoSummonMethods) SummonSingleton() *TwoSummonMethods {
	return new(TwoSummonMethods{})
}

func TestSummon(t *testing.T) {
	var _ = Binds[AnInterface, AnImplementation]()

	circle := NewCircle()

	tests := []struct {
		name      string
		wantType  reflect.Type
		wantErr   error
		runSummon func() (any, error)
	}{
		{
			name:      "Summon a leaf node",
			wantType:  reflect.TypeOf(LeafNode("")),
			runSummon: func() (any, error) { return circle.Summon[LeafNode]() },
		},
		{
			name:      "Summon an unregistered type fails",
			wantErr:   ErrNoSummoner,
			runSummon: func() (any, error) { return circle.Summon[string]() },
		},
		{
			name:      "Summon nested deps",
			wantType:  reflect.TypeOf(RootNode{}),
			runSummon: func() (any, error) { return circle.Summon[RootNode]() },
		},
		{
			name:      "Summon pointer",
			wantType:  reflect.TypeOf(new(Pointer{})),
			runSummon: func() (any, error) { return circle.Summon[*Pointer]() },
		},
		{
			name:      "Too many summoner methods fails",
			wantErr:   ErrInvalidSummonerMethod,
			runSummon: func() (any, error) { return circle.Summon[TwoSummonMethods]() },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.runSummon()

			if err != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error (%v), wanted (%v)", err, tc.wantErr)
				}
				return
			}

			gotType := reflect.TypeOf(got)

			if gotType != tc.wantType {
				t.Errorf("Summon returned type %s, %s expected", gotType, tc.wantType)
			}
		})
	}
}

func (p *Pointer) Summon() *Pointer {
	return new(Pointer{})
}

type Singleton string

func (s *Singleton) SummonSingleton() *Singleton {
	return new(Singleton(rand.Text()))
}

type InvalidSingleton struct{}

// Singletons should _always_ be pointers
func (s InvalidSingleton) SummonSingleton() InvalidSingleton {
	return InvalidSingleton{}
}

type SingletonWithDep struct {
	node LeafNode
}

func (s *SingletonWithDep) SummonSingleton(n LeafNode) *SingletonWithDep {
	return new(SingletonWithDep{n})
}

func TestSummon_Singleton(t *testing.T) {
	circle := NewCircle()

	tests := []struct {
		name      string
		wantType  reflect.Type
		wantErr   error
		runSummon func() (any, error)
	}{
		{
			name:      "Summon singleton",
			wantType:  reflect.TypeOf(new(Singleton(""))),
			runSummon: func() (any, error) { return circle.Summon[*Singleton]() },
		},
		{
			name:      "Summon singleton with dep",
			wantType:  reflect.TypeOf(new(SingletonWithDep{})),
			runSummon: func() (any, error) { return circle.Summon[*SingletonWithDep]() },
		},
		{
			name:      "Summon singleton fails if pointer not provided",
			wantErr:   ErrInvalidSummonerMethod,
			runSummon: func() (any, error) { return circle.Summon[InvalidSingleton]() },
		},
		{
			name:      "Summon singleton fails if requesting a value instead of a pointer",
			wantErr:   ErrInvalidSummonerMethod,
			runSummon: func() (any, error) { return circle.Summon[Singleton]() },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotFirst, _ := tc.runSummon()
			gotSecond, err := tc.runSummon()

			if err != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error (%v), wanted (%v)", err, tc.wantErr)
				}
				return
			}

			if gotFirst != gotSecond {
				t.Errorf("Summoned types are not identical: %v != %v", &gotFirst, &gotSecond)
			}

			gotType := reflect.TypeOf(gotFirst)
			if gotType != tc.wantType {
				t.Errorf("Summon returned type %s, %s expected", gotType, tc.wantType)
			}
		})
	}
}

func TestSummon_Singleton_ThreadSafe(t *testing.T) {
	threadtest.ResetCounters()
	circle := NewCircle()
	t.Run("Threadsafe summon singleton", func(t *testing.T) {
		const summons int = 100
		c := make(chan struct{})

		got := make([]*threadtest.LeafAtomicSingleton, summons)

		var ready, done sync.WaitGroup
		ready.Add(summons)
		done.Add(summons)
		for i := range summons {
			go func() {
				defer done.Done()
				ready.Done()

				// Wait for something to write to the channel
				<-c

				var err error
				got[i], err = circle.Summon[*threadtest.LeafAtomicSingleton]()
				if err != nil {
					t.Errorf("Got error (%v), expected none", err)
				}
			}()
		}

		ready.Wait()
		close(c)
		done.Wait()

		if threadtest.GetLeafCounter() != 1 {
			t.Errorf("Constructor for threadtest.LeafAtomicSingleton called %v times, expected 1.", threadtest.GetLeafCounter())
		}
	})
}

func TestSummon_NestedSingleton_ThreadSafe(t *testing.T) {
	threadtest.ResetCounters()
	circle := NewCircle()
	t.Run("Nested threadsafe summon singleton", func(t *testing.T) {
		const summons int = 100
		c := make(chan struct{})

		gotLeaf := make([]*threadtest.LeafAtomicSingleton, summons)
		gotRoot := make([]*threadtest.RootAtomicSingleton, summons)

		var ready, done sync.WaitGroup
		ready.Add(summons)
		done.Add(summons)
		for i := range summons {
			go func() {
				defer done.Done()
				ready.Done()

				// Wait for something to write to the channel
				<-c

				var err error
				if i%2 == 0 {
					gotLeaf[i], err = circle.Summon[*threadtest.LeafAtomicSingleton]()
				} else {
					gotRoot[i], err = circle.Summon[*threadtest.RootAtomicSingleton]()
				}
				if err != nil {
					t.Errorf("Got error (%v), expected none", err)
				}
			}()
		}

		ready.Wait()
		close(c)
		done.Wait()

		if threadtest.GetLeafCounter() != 1 {
			t.Errorf("Constructor for threadtest.LeafAtomicSingleton called %v times, expected 1.", threadtest.GetLeafCounter())
		}
		if threadtest.GetRootCounter() != 1 {
			t.Errorf("Constructor for threadtest.RootAtomicSingleton called %v times, expected 1.", threadtest.GetLeafCounter())
		}
	})
}

type A struct{}

func (a A) Summon(b B) A {
	return A{}
}

type B struct {
	a A
}

func (b B) Summon(_a A) B {
	return B{a: _a}
}

type APrime struct {
	a A
}

func (a APrime) Summon(_a A) APrime {
	return APrime{a: _a}
}

func TestSummon_CycleDetection(t *testing.T) {
	circle := NewCircle()

	tests := []struct {
		name      string
		wantErr   error
		runSummon func() (any, error)
	}{
		{
			name:      "Detect cycle",
			wantErr:   fmt.Errorf("%w: summon.A -> summon.B -> summon.A", ErrCircularDependency),
			runSummon: func() (any, error) { return circle.Summon[A]() },
		},
		{
			name:      "Detect cycle - ignoring types out of cycle",
			wantErr:   fmt.Errorf("%w: summon.A -> summon.B -> summon.A", ErrCircularDependency),
			runSummon: func() (any, error) { return circle.Summon[APrime]() },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a, err := tc.runSummon()

			if !errors.Is(err, ErrCircularDependency) {
				t.Errorf("Summoning type A should return nil, error - instead got %s, %s", a, err)
			}

			if err.Error() != tc.wantErr.Error() {
				t.Errorf("Incorrect error string in summoning type A, expected %s, got %s", tc.wantErr, err)
			}
		})
	}
}
