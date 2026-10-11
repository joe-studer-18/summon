package summon

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
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

func (p *Pointer) Summon() *Pointer {
	return new(Pointer{})
}

func TestSummon(t *testing.T) {
	var _ = Binds[AnInterface, AnImplementation]()

	circle := Circle{}

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
			wantErr:   ErrInvalidReturnType,
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
