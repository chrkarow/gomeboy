package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubRegister_GetValue(t *testing.T) {

	// GIVEN
	expectedVal := byte(0xFA)
	sr := newRegister().Hi
	sr.value = expectedVal

	// WHEN
	result := sr.GetValue()

	// THEN
	assert.Equal(t, expectedVal, result)
}

func TestSubRegister_SetValue(t *testing.T) {

	// GIVEN
	expectedVal := byte(0xFA)
	sr := newRegister().Hi

	// WHEN
	sr.SetValue(expectedVal)

	// THEN
	assert.Equal(t, expectedVal, sr.value)
}

func TestSubRegister_String(t *testing.T) {

	// GIVEN
	expectedVal := byte(0xFA)
	sr := newRegister().Hi
	sr.value = expectedVal

	// WHEN
	result := sr.String()

	// THEN
	assert.Equal(t, "0xFA", result)
}

func TestRegister_SetValue(t *testing.T) {

	// GIVEN
	r := newRegister()

	// WHEN
	r.SetValue(uint16(0xFAF0))

	// THEN
	assert.Equal(t, byte(0xFA), r.Hi.value)
	assert.Equal(t, byte(0xf0), r.Lo.value)
}

func TestRegister_GetValue(t *testing.T) {

	// GIVEN
	r := newRegister()
	r.Hi.value = byte(0xFA)
	r.Lo.value = byte(0xE2)

	// WHEN
	result := r.GetValue()

	// THEN
	assert.Equal(t, uint16(0xFAE2), result)
}

func TestRegister_String(t *testing.T) {

	// GIVEN
	expectedVal := uint16(0xFAE2)
	r := newRegister()
	r.SetValue(expectedVal)

	// WHEN
	result := r.String()
	// THEN
	assert.Equal(t, "0xFAE2", result)
}
