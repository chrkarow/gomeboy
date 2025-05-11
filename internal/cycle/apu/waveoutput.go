package apu

import "gameboy-emulator/internal/util"

const waveOutputLengthTimerMax = 256

type (
	WaveGenerator interface {
		Tick()
		LengthTick()
		Trigger()
		GetSample() byte
		IsEnabled() bool
		Disable()
	}

	WaveOutput struct {
		lengthCounter uint16
		lengthEnable  bool

		outputLevel byte
		volumeShift byte

		period uint
		dacOn  bool

		waveRAM [0x10]byte

		frequencyTimer uint
		samplePosition byte
		ticks          uint
		enabled        bool
		currentSample  byte
	}
)

func NewWaveOutput() *WaveOutput {
	return &WaveOutput{}
}

func (w *WaveOutput) Tick() {
	if !w.enabled {
		w.currentSample = 0x0
		return
	}

	if w.frequencyTimer--; w.frequencyTimer > 0 {
		return
	}
	w.resetFrequencyTimer()

	// Running through the bytes from top to bottom
	var sample byte
	if !w.dacOn {
		sample = 0x0
	} else if w.samplePosition%2 == 0 {
		sample = w.waveRAM[w.samplePosition/2] >> 4 // on even numbers take upper nibble
	} else {
		sample = w.waveRAM[w.samplePosition/2] & 0x0F // on odd numbers take lower nibble
	}

	w.currentSample = sample >> w.volumeShift

	w.samplePosition++
	if w.samplePosition == 0x20 {
		w.samplePosition = 0
	}
}

func (w *WaveOutput) LengthTick() {
	if !w.lengthEnable {
		return
	}
	w.lengthCounter--
	if w.lengthCounter == 0 {
		w.enabled = false
	}
}

func (w *WaveOutput) Trigger() {
	w.enabled = true
	w.samplePosition = 0
	w.currentSample = 0
	if w.lengthCounter == 0 {
		w.lengthCounter = waveOutputLengthTimerMax
	}
	w.resetFrequencyTimer()
}

func (w *WaveOutput) GetSample() byte {
	return w.currentSample
}

func (w *WaveOutput) IsEnabled() bool {
	return w.enabled
}

func (w *WaveOutput) Disable() {
	w.enabled = false
	w.currentSample = 0x0
	w.outputLevel = 0
	w.volumeShift = 0
	w.dacOn = false
	w.frequencyTimer = 0
	w.samplePosition = 0
	w.ticks = 0
}

func (w *WaveOutput) WriteWaveRAM(address byte, data byte) {
	w.waveRAM[address&0x0F] = data
}

func (w *WaveOutput) ReadWaveRAM(address byte) byte {
	return w.waveRAM[address&0x0F]
}

// SetNRx0 turns the DAC on or off
func (w *WaveOutput) SetNRx0(data byte) {
	if w.dacOn && !util.BitIsSet8(data, 7) {
		w.Disable()
	}
	w.dacOn = util.BitIsSet8(data, 7)
}

func (w *WaveOutput) GetNRx0() byte {
	if w.dacOn {
		return 0xFF
	}
	return 0x7F
}

// SetNRx1 sets the length timer
func (w *WaveOutput) SetNRx1(data byte) {
	w.lengthCounter = waveOutputLengthTimerMax - uint16(data)
}

func (w *WaveOutput) SetNRx2(data byte) {
	w.outputLevel = (data & 0x60) >> 5
	switch w.outputLevel {
	case 0:
		w.volumeShift = 4
	case 1:
		w.volumeShift = 0
	case 2:
		w.volumeShift = 1
	case 3:
		w.volumeShift = 2
	}
}

func (w *WaveOutput) GetNRx2() byte {
	return 0x9F | w.outputLevel<<5
}

func (w *WaveOutput) SetNRx3(data byte) {
	w.period &= 0xFF00
	w.period |= uint(data)
}

// SetNRx4 sets the high bits of the period, enables the length timer and triggers the channel
func (w *WaveOutput) SetNRx4(data byte) {

	if util.BitIsSet8(data, 7) {
		w.Trigger()
	}

	w.lengthEnable = util.BitIsSet8(data, 6)

	w.period &= 0x00FF
	w.period |= uint(data&0x7) << 8
}

func (w *WaveOutput) GetNRx4() byte {
	if w.lengthEnable {
		return 0xFF
	}
	return 0xBF
}

func (w *WaveOutput) resetFrequencyTimer() {
	w.frequencyTimer = (2048 - w.period) * 2
}
