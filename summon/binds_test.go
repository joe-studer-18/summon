package summon

import (
	"errors"
	"reflect"
	"testing"
)

type BindsAnInterface interface {
	Do() string
}

type BindsAnImplementation struct{}

func (a BindsAnImplementation) Do() string {
	return "BindsAnImplementation"
}

func (a BindsAnImplementation) DiGoProvide() BindsAnImplementation {
	return BindsAnImplementation{}
}

type NestedDepOnInterface struct{}

func (n NestedDepOnInterface) DiGoProvide(i BindsAnInterface) NestedDepOnInterface {
	return NestedDepOnInterface{}
}

func TestBinds(t *testing.T) {
	inj := Injector{}

	tests := []struct {
		name     string
		runBind  func() error
		runGet   func() (any, error)
		wantType reflect.Type
		wantErr  error
	}{
		{
			name:     "ok",
			runBind:  func() error { return Binds[BindsAnInterface, BindsAnImplementation]() },
			runGet:   func() (any, error) { return inj.Get[BindsAnInterface]() },
			wantType: reflect.TypeOf(BindsAnImplementation{}),
		},
		{
			name:     "Bind nested dep on an interface",
			runBind:  func() error { return Binds[BindsAnInterface, BindsAnImplementation]() },
			runGet:   func() (any, error) { return inj.Get[NestedDepOnInterface]() },
			wantType: reflect.TypeOf(NestedDepOnInterface{}),
		},
		{
			name: "Bind fails on collision",
			runBind: func() error {
				err := Binds[BindsAnInterface, BindsAnImplementation]()
				if err != nil {
					return err
				}
				return Binds[BindsAnInterface, string]()
			},
			runGet:  func() (any, error) { return inj.Get[NestedDepOnInterface]() },
			wantErr: ErrBindCollision,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.runBind()
			if err != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("got error (%v), wanted (%v)", err, tc.wantErr)
				}
			} else {
				got, err := tc.runGet()

				if err != nil {
					if !errors.Is(err, tc.wantErr) {
						t.Errorf("got error (%v), wanted (%v)", err, tc.wantErr)
					}
				} else {
					gotType := reflect.TypeOf(got)

					if gotType != tc.wantType {
						t.Errorf("Get returned type %s, %s expected", gotType, tc.wantType)
					}
				}
			}
			// Clean up bindMap
			for t2 := range bindMap {
				delete(bindMap, t2)
			}
		})
	}
}
