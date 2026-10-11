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

func (a AnImplementation) DiGoProvide() AnImplementation {
	return AnImplementation{}
}

type RootNode struct {
	name string
}

func (f RootNode) Process() bool {
	fmt.Printf("Foo: %s\n", f.name)
	return true
}

func (f RootNode) DiGoProvide(name LeafNode, i AnInterface) RootNode {
	return RootNode{fmt.Sprintf("%s.%s", string(name), i.Do())}
}

type LeafNode string

func (f LeafNode) DiGoProvide() LeafNode {
	return "DiGo"
}

type Pointer struct{}

func (p *Pointer) DiGoProvide() *Pointer {
	return new(Pointer{})
}

func TestGet(t *testing.T) {
	var _ = Binds[AnInterface, AnImplementation]()

	inj := Injector{}

	tests := []struct {
		name     string
		wantType reflect.Type
		wantErr  error
		runGet   func() (any, error)
	}{
		{
			name:     "Get a leaf node",
			wantType: reflect.TypeOf(LeafNode("")),
			runGet:   func() (any, error) { return inj.Get[LeafNode]() },
		},
		// {
		// 	name:    "Get an unregistered type fails",
		// 	wantErr: ErrInvalidReturnType,
		// 	runGet:  func() (any, error) { return inj.Get[string]() },
		// },
		{
			name:     "Get nested deps",
			wantType: reflect.TypeOf(RootNode{}),
			runGet:   func() (any, error) { return inj.Get[RootNode]() },
		},
		{
			name:     "Get pointer",
			wantType: reflect.TypeOf(new(Pointer{})),
			runGet:   func() (any, error) { return inj.Get[*Pointer]() },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.runGet()

			if err != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error (%v), wanted (%v)", err, tc.wantErr)
				}
				return
			}

			gotType := reflect.TypeOf(got)

			if gotType != tc.wantType {
				t.Errorf("Get returned type %s, %s expected", gotType, tc.wantType)
			}
		})
	}
}
