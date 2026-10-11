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

func (a BindsAnImplementation) Summon() BindsAnImplementation {
	return BindsAnImplementation{}
}

type NestedDepOnInterface struct{}

func (n NestedDepOnInterface) Summon(i BindsAnInterface) NestedDepOnInterface {
	return NestedDepOnInterface{}
}

func TestBinds(t *testing.T) {
	circle := Circle{}

	tests := []struct {
		name      string
		runBind   func() error
		runSummon func() (any, error)
		wantType  reflect.Type
		wantErr   error
	}{
		{
			name:      "ok",
			runBind:   func() error { return Binds[BindsAnInterface, BindsAnImplementation]() },
			runSummon: func() (any, error) { return circle.Summon[BindsAnInterface]() },
			wantType:  reflect.TypeOf(BindsAnImplementation{}),
		},
		{
			name:      "Bind nested dep on an interface",
			runBind:   func() error { return Binds[BindsAnInterface, BindsAnImplementation]() },
			runSummon: func() (any, error) { return circle.Summon[NestedDepOnInterface]() },
			wantType:  reflect.TypeOf(NestedDepOnInterface{}),
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
			runSummon: func() (any, error) { return circle.Summon[NestedDepOnInterface]() },
			wantErr:   ErrBindCollision,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.runBind()
			if err != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("summoned error (%v), wanted (%v)", err, tc.wantErr)
				}
			} else {
				summoned, err := tc.runSummon()

				if err != nil {
					if !errors.Is(err, tc.wantErr) {
						t.Errorf("summoned error (%v), wanted (%v)", err, tc.wantErr)
					}
				} else {
					summonedType := reflect.TypeOf(summoned)

					if summonedType != tc.wantType {
						t.Errorf("Summon returned type %s, %s expected", summonedType, tc.wantType)
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
