package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSet(t *testing.T) {
	// GIVEN
	value := byte(0)

	// WHEN
	SetBit(&value, 1)

	// THEN
	assert.Equal(t, byte(2), value)
}

func TestUnsetSet(t *testing.T) {
	// GIVEN
	value := byte(0xFF)

	// WHEN
	UnsetBit8(&value, 3)

	// THEN
	assert.Equal(t, byte(0xF7), value)
}
