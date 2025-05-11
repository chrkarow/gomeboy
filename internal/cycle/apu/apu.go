package apu

import "gameboy-emulator/internal/util"

const GameBoyClockSpeed uint = 4 * 1024 * 1024
const SamplingRate = 44100

type (
	APU struct {
		channel1 *SweepableSquareWave
		channel2 *SquareWave
		channel3 *WaveOutput
		channel4 *Noise

		frameSequencer *FrameSequencer
		ticks          uint

		panning     byte
		volumeLeft  byte
		volumeRight byte
		vinLeft     bool // unused
		vinRight    bool // unused
		enabled     bool
	}
)

func New() *APU {
	a := &APU{}
	a.Reset()
	return a
}

func (a *APU) Reset() {
	a.channel1 = NewSweepableSquareWave(NewSquareWave())
	a.channel2 = NewSquareWave()
	a.channel3 = NewWaveOutput()
	a.channel4 = NewNoise()
	a.frameSequencer = NewFrameSequencer(a.channel1, a.channel2, a.channel3, a.channel4)
	a.enabled = false
	a.ticks = 0
}

func (a *APU) Tick() (left byte, right byte, play bool) {
	a.frameSequencer.Tick() // Has to keep ticking to stay in sync with DIV

	if !a.enabled {
		return
	}

	if a.ticks++; a.ticks < GameBoyClockSpeed/SamplingRate {
		return
	}
	a.ticks = 0

	leftRaw, rightRaw := a.panAndMix(
		a.channel1.GetSample(),
		a.channel2.GetSample(),
		a.channel3.GetSample(),
		a.channel4.GetSample(),
	)

	left = (a.volumeLeft + 1) * leftRaw
	right = (a.volumeRight + 1) * rightRaw
	play = true

	return
}

func (a *APU) panAndMix(c1Sample byte, c2Sample byte, c3Sample byte, c4Sample byte) (left byte, right byte) {
	var leftRaw [4]byte
	var rightRaw [4]byte

	if util.BitIsSet8(a.panning, 0) {
		rightRaw[0] = c1Sample
	}

	if util.BitIsSet8(a.panning, 1) {
		rightRaw[1] = c2Sample
	}

	if util.BitIsSet8(a.panning, 2) {
		rightRaw[2] = c3Sample
	}

	if util.BitIsSet8(a.panning, 3) {
		rightRaw[3] = c4Sample
	}

	if util.BitIsSet8(a.panning, 4) {
		leftRaw[0] = c1Sample
	}

	if util.BitIsSet8(a.panning, 5) {
		leftRaw[1] = c2Sample
	}

	if util.BitIsSet8(a.panning, 6) {
		leftRaw[2] = c3Sample
	}

	if util.BitIsSet8(a.panning, 7) {
		leftRaw[3] = c4Sample
	}

	return byte(util.Sum(leftRaw[:]) / 4), byte(util.Sum(rightRaw[:]) / 4)
}

// Channel 1 ##########################

func (a *APU) WriteNR10(data byte) {
	if !a.enabled {
		return
	}
	a.channel1.SetNRx0(data)
}

func (a *APU) ReadNR10() byte {
	return a.channel1.GetNRx0()
}

func (a *APU) WriteNR11(data byte) {
	if !a.enabled {
		return
	}
	a.channel1.SetNRx1(data)
}

func (a *APU) ReadNR11() byte {
	return a.channel1.GetNRx1()
}

func (a *APU) WriteNR12(data byte) {
	if !a.enabled {
		return
	}
	a.channel1.SetNRx2(data)
}

func (a *APU) ReadNR12() byte {
	return a.channel1.GetNRx2()
}

func (a *APU) WriteNR13(data byte) {
	if !a.enabled {
		return
	}
	a.channel1.SetNRx3(data)
}

func (a *APU) ReadNR13() byte {
	return 0xFF
}

func (a *APU) WriteNR14(data byte) {
	if !a.enabled {
		return
	}
	a.channel1.SetNRx4(data)
}

func (a *APU) ReadNR14() byte {
	return a.channel1.GetNRx4()
}

// Channel 2 ##########################

func (a *APU) WriteNR21(data byte) {
	if !a.enabled {
		return
	}
	a.channel2.SetNRx1(data)
}

func (a *APU) ReadNR21() byte {
	return a.channel2.GetNRx1()
}

func (a *APU) WriteNR22(data byte) {
	if !a.enabled {
		return
	}
	a.channel2.SetNRx2(data)
}

func (a *APU) ReadNR22() byte {
	return a.channel2.GetNRx2()
}

func (a *APU) WriteNR23(data byte) {
	if !a.enabled {
		return
	}
	a.channel2.SetNRx3(data)
}

func (a *APU) ReadNR23() byte {
	return 0xFF
}

func (a *APU) WriteNR24(data byte) {
	if !a.enabled {
		return
	}
	a.channel2.SetNRx4(data)
}

func (a *APU) ReadNR24() byte {
	return a.channel2.GetNRx4()
}

// Channel 3 ##########################

func (a *APU) WriteNR30(data byte) {
	if !a.enabled {
		return
	}
	a.channel3.SetNRx0(data)
}

func (a *APU) ReadNR30() byte {
	return a.channel3.GetNRx0()
}

func (a *APU) WriteNR31(data byte) {
	if !a.enabled {
		return
	}
	a.channel3.SetNRx1(data)
}

func (a *APU) ReadNR31() byte {
	return 0xFF
}

func (a *APU) WriteNR32(data byte) {
	if !a.enabled {
		return
	}
	a.channel3.SetNRx2(data)
}

func (a *APU) ReadNR32() byte {
	return a.channel3.GetNRx2()
}

func (a *APU) WriteNR33(data byte) {
	if !a.enabled {
		return
	}
	a.channel3.SetNRx3(data)
}

func (a *APU) ReadNR33() byte {
	return 0xFF
}

func (a *APU) WriteNR34(data byte) {
	if !a.enabled {
		return
	}
	a.channel3.SetNRx4(data)
}

func (a *APU) ReadNR34() byte {
	return a.channel3.GetNRx4()
}

// Channel 4 ##########################

func (a *APU) WriteWaveRAM(address byte, data byte) {
	a.channel3.WriteWaveRAM(address, data)
}

func (a *APU) ReadWaveRAM(address byte) byte {
	return a.channel3.ReadWaveRAM(address)
}

func (a *APU) WriteNR41(data byte) {
	if !a.enabled {
		return
	}
	a.channel4.SetNRx1(data)
}

func (a *APU) ReadNR41() byte {
	return 0xFF
}

func (a *APU) WriteNR42(data byte) {
	if !a.enabled {
		return
	}
	a.channel4.SetNRx2(data)
}

func (a *APU) ReadNR42() byte {
	return a.channel4.GetNRx2()
}

func (a *APU) WriteNR43(data byte) {
	if !a.enabled {
		return
	}
	a.channel4.SetNRx3(data)
}

func (a *APU) ReadNR43() byte {
	return a.channel4.GetNRx3()
}

func (a *APU) WriteNR44(data byte) {
	if !a.enabled {
		return
	}
	a.channel4.SetNRx4(data)
}

func (a *APU) ReadNR44() byte {
	return a.channel4.GetNRx4()
}

// Control Registers #######################

func (a *APU) WriteNR50(data byte) {
	if !a.enabled {
		return
	}
	a.vinLeft = util.BitIsSet8(data, 7)
	a.volumeLeft = (data & 0x70) >> 4
	a.vinRight = util.BitIsSet8(data, 3)
	a.volumeRight = data & 0x7
}

func (a *APU) ReadNR50() byte {
	var vinLeftBit byte
	if a.vinLeft {
		vinLeftBit = 0x80
	}

	var vinRightBit byte
	if a.vinRight {
		vinRightBit = 0x8
	}

	return vinLeftBit | a.volumeLeft<<4 | vinRightBit | a.volumeRight
}

func (a *APU) WriteNR51(data byte) {
	if !a.enabled {
		return
	}
	a.panning = data
}

func (a *APU) ReadNR51() byte {
	return a.panning
}

func (a *APU) WriteNR52(data byte) {
	if a.enabled && !util.BitIsSet8(data, 7) {
		a.clearRegisters()
	}
	a.enabled = util.BitIsSet8(data, 7)
	a.frameSequencer.SetEnabled(a.enabled)
}

func (a *APU) ReadNR52() byte {
	var apuEnabledBit byte
	if a.enabled {
		apuEnabledBit = 0x80
	}
	var ch4EnabledBit byte
	if a.channel4.IsEnabled() {
		ch4EnabledBit = 0x8
	}
	var ch3EnabledBit byte
	if a.channel3.IsEnabled() {
		ch3EnabledBit = 0x4
	}
	var ch2EnabledBit byte
	if a.channel2.IsEnabled() {
		ch2EnabledBit = 0x2
	}
	var ch1EnabledBit byte
	if a.channel1.IsEnabled() {
		ch1EnabledBit = 0x1
	}

	return apuEnabledBit | 0x70 | ch4EnabledBit | ch3EnabledBit | ch2EnabledBit | ch1EnabledBit
}

func (a *APU) clearRegisters() {
	a.WriteNR10(0x0)
	a.WriteNR11(0x0)
	a.WriteNR12(0x0)
	a.WriteNR13(0x0)
	a.WriteNR14(0x0)
	a.WriteNR21(0x0)
	a.WriteNR22(0x0)
	a.WriteNR23(0x0)
	a.WriteNR24(0x0)
	a.WriteNR30(0x0)
	a.WriteNR31(0x0)
	a.WriteNR32(0x0)
	a.WriteNR33(0x0)
	a.WriteNR34(0x0)
	a.WriteNR41(0x0)
	a.WriteNR42(0x0)
	a.WriteNR43(0x0)
	a.WriteNR44(0x0)
	a.WriteNR50(0x0)
	a.WriteNR51(0x0)
}
