package digo

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type RootNode struct {
	name string
}

func (f RootNode) Process() bool {
	fmt.Printf("Foo: %s\n", f.name)
	return true
}

func (f RootNode) DiGoProvide(name LeafNode) RootNode {
	return RootNode{string(name)}
}

type LeafNode string

func (f LeafNode) DiGoProvide() LeafNode {
	return "DiGo"
}

func TestGet(t *testing.T) {
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
		{
			name:    "Get an unregistered type fails",
			wantErr: ErrInvalidReturnType,
			runGet:  func() (any, error) { return inj.Get[string]() },
		},
		{
			name:     "Get nested deps",
			wantType: reflect.TypeOf(RootNode{}),
			runGet:   func() (any, error) { return inj.Get[RootNode]() },
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
