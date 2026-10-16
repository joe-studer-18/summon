//go:build unit

package summon

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
		wantPanic error
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
			wantPanic: ErrBindCollision,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic != nil {
				assert.Panics(t, func() { tc.runBind() })
				return
			}
			err := tc.runBind()
			if err != nil {
				t.Errorf("got error (%v), wanted (%v)", err, nil)
			} else {
				summoned, err := tc.runSummon()

				if err != nil {
					t.Errorf("got error (%v), wanted (%v)", err, nil)
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
