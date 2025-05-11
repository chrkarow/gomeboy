package apu

import "gameboy-emulator/internal/util"

const squareWaveLengthTimerMax = 64

type (
	Enveloper interface {
		WaveGenerator
		EnvelopeTick()
	}

	SquareWave struct {
		lengthCounter byte
		lengthEnable  bool

		volumeEnvelope *VolumeEnvelope

		period         uint
		dutyCycleIndex byte

		frequencyTimer uint
		dutyPosition   byte
		ticks          uint
		enabled        bool
		currentSample  byte
	}
)

// DutyCycles represents available duty patterns. For any given frequency,
// we'll internally split one period of that frequency in 8, and for each
// of those slices, this will specify whether the signal should be on or off.
var dutyCycles = [4][8]bool{
	{false, false, false, false, false, false, false, true}, // 00000001, 12.5%
	{true, false, false, false, false, false, false, true},  // 10000001, 25%
	{true, false, false, false, false, true, true, true},    // 10000111, 50%
	{false, true, true, true, true, true, true, false},      // 01111110, 75%
}

func NewSquareWave() *SquareWave {
	return &SquareWave{
		volumeEnvelope: NewVolumeEnvelope(),
	}
}

func (sq *SquareWave) Tick() {
	if !sq.enabled {
		sq.currentSample = 0x0
		return
	}

	if sq.frequencyTimer--; sq.frequencyTimer > 0 {
		return
	}
	sq.resetFrequencyTimer()
	sq.dutyPosition = (sq.dutyPosition + 1) % 8

	if dutyCycles[sq.dutyCycleIndex][sq.dutyPosition] {
		sq.currentSample = sq.volumeEnvelope.GetVolume()
	} else {
		sq.currentSample = 0x0
	}
}

func (sq *SquareWave) LengthTick() {
	if !sq.lengthEnable {
		return
	}
	sq.lengthCounter--
	if sq.lengthCounter == 0 {
		sq.Disable()
	}
}

func (sq *SquareWave) EnvelopeTick() {
	sq.volumeEnvelope.Tick()
}

func (sq *SquareWave) Trigger() {
	sq.enabled = true
	sq.dutyPosition = 0
	sq.currentSample = 0
	sq.resetFrequencyTimer()
	if sq.lengthCounter == 0 {
		sq.lengthCounter = squareWaveLengthTimerMax
	}
	sq.volumeEnvelope.Trigger()
}

func (sq *SquareWave) GetSample() byte {
	return sq.currentSample
}

func (sq *SquareWave) IsEnabled() bool {
	return sq.enabled
}

func (sq *SquareWave) Disable() {
	sq.volumeEnvelope.Disable()
	sq.enabled = false
	sq.currentSample = 0x0
	sq.dutyCycleIndex = 0
	sq.dutyPosition = 0
	sq.period = 0
	sq.frequencyTimer = 0
	sq.ticks = 0
}

func (sq *SquareWave) resetFrequencyTimer() {
	sq.frequencyTimer = (2048 - sq.period) * 4
}

// SetNRx1 sets the length timer and duty cycle
func (sq *SquareWave) SetNRx1(data byte) {
	sq.dutyCycleIndex = data >> 6

	sq.lengthCounter = squareWaveLengthTimerMax - (data & 0x3F)
}

// GetNRx1 returns the value of the NRx1 register. only bit 6 and 7 are readable. All others set to 1.
func (sq *SquareWave) GetNRx1() byte {
	return (sq.dutyCycleIndex << 6) | 0x3F
}

// SetNRx2 controls the volume envelope of this channel
func (sq *SquareWave) SetNRx2(data byte) {
	sq.volumeEnvelope.Write(data)
	if !sq.volumeEnvelope.IsEnabled() {
		sq.Disable()
	}
}

// GetNRx2 returns the value of the NRx2 register.
func (sq *SquareWave) GetNRx2() byte {
	return sq.volumeEnvelope.Read()
}

// SetNRx3 sets the low bits of the period
func (sq *SquareWave) SetNRx3(data byte) {
	sq.period &= 0xFF00
	sq.period |= uint(data)
}

// SetNRx4 sets the high bits of the period, enables the length timer and triggers the channel
func (sq *SquareWave) SetNRx4(data byte) {

	if util.BitIsSet8(data, 7) {
		sq.Trigger()
	}

	sq.lengthEnable = util.BitIsSet8(data, 6)

	sq.period &= 0x00FF
	sq.period |= uint(data&0x7) << 8
}

func (sq *SquareWave) GetNRx4() byte {
	if sq.lengthEnable {
		return 0xFF
	}
	return 0xBF
}
