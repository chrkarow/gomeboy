package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetFlag(t *testing.T) {

	// GIVEN
	f := flags{0x00}

	// WHEN
	f.setFlag(n)

	// THEN
	assert.Equal(t, byte(0b01000000), f.value)
}

func TestUnsetFlag(t *testing.T) {

	// GIVEN
	f := flags{0xF0}

	// WHEN
	f.unsetFlag(h)

	// THEN
	assert.Equal(t, byte(0b11010000), f.value)
}

func TestIsSet(t *testing.T) {

	// GIVEN
	f := flags{0b00010000}

	// WHEN
	result := f.isSet(c)
	result2 := f.isSet(z)

	// THEN
	assert.True(t, result)
	assert.False(t, result2)
}
