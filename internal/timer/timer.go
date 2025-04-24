package timer

import (
	"gameboy-emulator/internal/bit"
	"gameboy-emulator/internal/interrupts"
)

const dividerThreshold int = 256

// Timer represents the timer and divider registers which are updated in certain frequencies.
//
// Source: https://gbdev.io/pandocs/Timer_and_Divider_Registers.html
type Timer struct {
	timerCounter   int
	dividerCounter int

	tima byte // Timer counter
	tma  byte // Timer modulo
	tac  byte // Timer control
	div  byte // Divider

	interrupts *interrupts.Interrupts
}

func New(inter *interrupts.Interrupts) *Timer {
	t := &Timer{
		interrupts: inter,
	}
	t.Reset()
	return t
}

// Reset the timer to initial state.
//
// Values taken from https://github.com/Gekkio/mooneye-test-suite/blob/main/acceptance/boot_hwio-dmgABCmgb.s
func (t *Timer) Reset() {
	t.tima = 0x0
	t.tma = 0x0
	t.tac = 0xF8
	t.div = 0xAD

	t.timerCounter = 0
	t.dividerCounter = 0
}

func (t *Timer) UpdateTimer(stepCycles int) {

	t.doUpdateDivider(stepCycles)

	if !t.isTimerEnabled() {
		return
	}

	t.doUpdateTimer(stepCycles)
}

func (t *Timer) GetDiv() byte {
	return t.div
}

func (t *Timer) ResetDiv() {
	t.div = 0x00
}

func (t *Timer) GetTima() byte {
	return t.tima
}

func (t *Timer) SetTima(value byte) {
	t.tima = value
}

func (t *Timer) GetTma() byte {
	return t.tma
}

func (t *Timer) SetTma(value byte) {
	t.tma = value
}

func (t *Timer) GetTac() byte {
	return t.tac
}

func (t *Timer) SetTac(value byte) {
	t.tac = value
}

// doUpdateDivider updates the value of div at a rate of 16384 Hz.
// This means that every dividerThreshold cycles the value is incremented.
func (t *Timer) doUpdateDivider(cyclesSinceLastUpdate int) {
	t.dividerCounter += cyclesSinceLastUpdate

	if t.dividerCounter < dividerThreshold {
		return
	}

	t.dividerCounter %= dividerThreshold
	t.div++

}

func (t *Timer) doUpdateTimer(cyclesSinceLastUpdate int) {
	t.timerCounter += cyclesSinceLastUpdate

	updateThreshold := t.getUpdateThreshold()
	if t.timerCounter < updateThreshold {
		return
	}

	t.timerCounter %= updateThreshold

	if t.tima == 0xFF {
		t.tima = t.tma
		t.interrupts.RequestInterrupt(interrupts.Timer)
	} else {
		t.tima++
	}
}

func (t *Timer) getUpdateThreshold() int {
	switch t.tac & 0x3 {
	case 0:
		return 1024 // = Clock speed (4194304 Hz) / 4096 Hz
	case 1:
		return 16 // = Clock speed (4194304 Hz) / 262144 Hz
	case 2:
		return 64 // = Clock speed (4194304 Hz) / 65536 Hz
	case 3:
		return 256 // = Clock speed (4194304 Hz) / 16384 Hz
	}
	return 0
}

func (t *Timer) isTimerEnabled() bool {
	return bit.IsSet8(t.tac, 2) // bit 2 has to be set
}
