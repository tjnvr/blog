package validation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeValidator struct {
	errs []error
}

func (f fakeValidator) Validate([]byte) []error { return f.errs }

func TestNewRegistryWithValidators_ShouldHoldGivenValidators(t *testing.T) {
	// setup
	registry := NewRegistryWithValidators(fakeValidator{}, fakeValidator{})

	// test
	got := len(registry.validators)

	// expect
	assert.Equal(t, 2, got)
}

func TestRegistry_Register_ShouldAppendValidator(t *testing.T) {
	// setup
	registry := NewRegistryWithValidators()

	// test
	registry.Register(fakeValidator{})

	// expect
	assert.Len(t, registry.validators, 1)
}

func TestRegistry_Validate_ShouldReturnNilWhenAllValidatorsPass(t *testing.T) {
	// setup
	registry := NewRegistryWithValidators(
		fakeValidator{errs: nil},
		fakeValidator{errs: []error{}},
	)

	// test
	err := registry.Validate([]byte("<html></html>"))

	// expect
	assert.NoError(t, err)
}

func TestRegistry_Validate_ShouldJoinErrorsFromEveryValidator(t *testing.T) {
	// given
	first := errors.New("error 1")
	second := errors.New("error 2")
	third := errors.New("error 3")

	// setup
	registry := NewRegistryWithValidators(
		fakeValidator{errs: []error{first}},
		fakeValidator{errs: []error{second, third}},
	)

	// test
	err := registry.Validate([]byte("<html></html>"))

	// expect
	assert.ErrorIs(t, err, first)
	assert.ErrorIs(t, err, second)
	assert.ErrorIs(t, err, third)
}

func TestRegistry_Validate_ShouldReturnNilWhenNoValidatorsRegistered(t *testing.T) {
	// setup
	registry := NewRegistryWithValidators()

	// test
	err := registry.Validate([]byte("<html></html>"))

	// expect
	assert.NoError(t, err)
}
