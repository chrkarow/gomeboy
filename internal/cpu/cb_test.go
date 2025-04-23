package cpu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShiftRight_logical_carryFlagSet(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10011011)

	// WHEN
	result := shiftRight(value, &f, false)

	// THEN
	assert.Equal(t, byte(0b00010000), f.getValue()) // only c flag set
	assert.Equal(t, byte(0b01001101), result)
}

func TestShiftRight_arithmetical_carryFlagSet(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10011011)

	// WHEN
	result := shiftRight(value, &f, true)

	// THEN
	assert.Equal(t, byte(0b00010000), f.getValue()) // only c flag set
	assert.Equal(t, byte(0b11001101), result)
}

func TestShiftRight_arithmetical_carryFlagNotSet(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10011010)

	// WHEN
	result := shiftRight(value, &f, true)

	// THEN
	assert.Equal(t, byte(0x00), f.getValue()) // no flags set
	assert.Equal(t, byte(0b11001101), result)
}

func TestShiftRight_arithmetical_resultIsZero(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b00000001)

	// WHEN
	result := shiftRight(value, &f, true)

	// THEN
	assert.Equal(t, byte(0b10010000), f.getValue()) // c and z flag set
	assert.Equal(t, byte(0b00000000), result)
}

func TestShiftLeft_carryFlagSet(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10011011)

	// WHEN
	result := shiftLeft(value, &f)

	// THEN
	assert.Equal(t, byte(0b00010000), f.getValue()) // only c flag set
	assert.Equal(t, byte(0b00110110), result)
}

func TestShiftLeft_carryFlagNotSet(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b00011010)

	// WHEN
	result := shiftLeft(value, &f)

	// THEN
	assert.Equal(t, byte(0x00), f.getValue()) // no flags set
	assert.Equal(t, byte(0b00110100), result)
}

func TestShiftLeft_resultIsZero(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10000000)

	// WHEN
	result := shiftLeft(value, &f)

	// THEN
	assert.Equal(t, byte(0b10010000), f.getValue()) // c and z flag set
	assert.Equal(t, byte(0b00000000), result)
}

func TestRotateLeft_carryFlagSet(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10011011)

	// WHEN
	result := rotateLeft(value, &f)

	// THEN
	assert.Equal(t, byte(0b00010000), f.getValue()) // only c flag set
	assert.Equal(t, byte(0b00110111), result)
}

func TestRotateLeft_carryFlagNotSetAfterwards(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b00011010)

	// WHEN
	result := rotateLeft(value, &f)

	// THEN
	assert.Equal(t, byte(0x00), f.getValue()) // no flags set
	assert.Equal(t, byte(0b00110100), result)
}

func TestRotateLeft_resultIsZero(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b00000000)

	// WHEN
	result := rotateLeft(value, &f)

	// THEN
	assert.Equal(t, byte(0b10000000), f.getValue()) //  z flag set
	assert.Equal(t, byte(0b00000000), result)
}

func TestRotateLeftThroughCarry_carryFlagSetAfterwards(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10011011)

	// WHEN
	result := rotateLeftThroughCarry(value, &f)

	// THEN
	assert.Equal(t, byte(0b00010000), f.getValue()) // only c flag set
	assert.Equal(t, byte(0b00110110), result)
}

func TestRotateLeftThroughCarry_carryFlagNotSetAfterwards(t *testing.T) {

	// GIVEN
	f := flags{0b00010000}
	value := byte(0b00011011)

	// WHEN
	result := rotateLeftThroughCarry(value, &f)

	// THEN
	assert.Equal(t, byte(0x00), f.getValue()) // no flags set
	assert.Equal(t, byte(0b00110111), result)
}

func TestRotateLeftThroughCarry_resultIsZero(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10000000)

	// WHEN
	result := rotateLeftThroughCarry(value, &f)

	// THEN
	assert.Equal(t, byte(0b10010000), f.getValue()) //  z and c flags set
	assert.Equal(t, byte(0b00000000), result)
}

func TestRotateRight_carryFlagSetAfterwards(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10011011)

	// WHEN
	result := rotateRight(value, &f)

	// THEN
	assert.Equal(t, byte(0b00010000), f.getValue()) // only c flag set
	assert.Equal(t, byte(0b11001101), result)
}

func TestRotateRight_carryFlagNotSetAfterwards(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b00011010)

	// WHEN
	result := rotateRight(value, &f)

	// THEN
	assert.Equal(t, byte(0x00), f.getValue()) // no flags set
	assert.Equal(t, byte(0b00001101), result)
}

func TestRotateRight_resultIsZero(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b00000000)

	// WHEN
	result := rotateRight(value, &f)

	// THEN
	assert.Equal(t, byte(0b10000000), f.getValue()) //  z flag set
	assert.Equal(t, byte(0b00000000), result)
}

func TestRotateRightThroughCarry_carryFlagSetAfterwards(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10011011)

	// WHEN
	result := rotateRightThroughCarry(value, &f)

	// THEN
	assert.Equal(t, byte(0b00010000), f.getValue()) // only c flag set
	assert.Equal(t, byte(0b01001101), result)
}

func TestRotateRightThroughCarry_carryFlagNotSetAfterwards(t *testing.T) {

	// GIVEN
	f := flags{0b00010000}
	value := byte(0b00011010)

	// WHEN
	result := rotateRightThroughCarry(value, &f)

	// THEN
	assert.Equal(t, byte(0x00), f.getValue()) // no flags set
	assert.Equal(t, byte(0b10001101), result)
}

func TestRotateRightThroughCarry_resultIsZero(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b00000001)

	// WHEN
	result := rotateRightThroughCarry(value, &f)

	// THEN
	assert.Equal(t, byte(0b10010000), f.getValue()) //  z and c flags set
	assert.Equal(t, byte(0b00000000), result)
}

func TestSwap(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b10111101)

	// WHEN
	result := swap(value, &f)

	// THEN
	assert.Equal(t, byte(0x00), f.getValue()) // no flags set
	assert.Equal(t, byte(0b11011011), result)
}

func TestSwap_resultIsZero(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b00000000)

	// WHEN
	result := swap(value, &f)

	// THEN
	assert.Equal(t, byte(0b10000000), f.getValue()) // no flags set
	assert.Equal(t, byte(0b00000000), result)
}

func TestBit_checkedBitIsSet(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b11001010)
	index := byte(6)

	// WHEN
	bit(value, index, &f)

	// THEN
	assert.Equal(t, byte(0b00100000), f.getValue()) // only h flag set
}

func TestBit_checkedBitIsNotSet(t *testing.T) {

	// GIVEN
	f := flags{0x00}
	value := byte(0b11001010)
	index := byte(0)

	// WHEN
	bit(value, index, &f)

	// THEN
	assert.Equal(t, byte(0b10100000), f.getValue()) // z and h flag set
}

func TestRes(t *testing.T) {

	// GIVEN
	value := byte(0b11111111)
	index := byte(2)

	// WHEN
	result := res(value, index)

	// THEN
	assert.Equal(t, byte(0b11111011), result)
}

func TestSet(t *testing.T) {

	// GIVEN
	value := byte(0b10000000)
	index := byte(4)

	// WHEN
	result := set(value, index)

	// THEN
	assert.Equal(t, byte(0b10010000), result)
}
