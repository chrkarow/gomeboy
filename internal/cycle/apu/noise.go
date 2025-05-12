package apu

import "gameboy-emulator/internal/util"

const noiseLengthTimerMax = 64

type Noise struct {
	lfsr uint16

	lengthCounter byte
	lengthEnable  bool

	frequencyTimer uint
	clockShift     byte
	use7Bit        bool
	clockDivider   byte

	volumeEnvelope *VolumeEnvelope
	currentSample  byte
	enabled        bool
}

func NewNoise() *Noise {
	return &Noise{
		volumeEnvelope: NewVolumeEnvelope(),
	}
}

func (n *Noise) Tick() {
	if !n.enabled {
		n.currentSample = 0x0
		return
	}

	if n.frequencyTimer--; n.frequencyTimer > 0 {
		return
	}

	//Whenever the frequency timer expires the following operations take place,
	//
	// 1. The frequency timer is reloaded using the above formula.
	// 2. The XNOR result of the 0th and 1st bit of LFSR is computed.
	// 3. The LFSR is shifted right by one bit and the above XNOR result is stored in bit 14.
	// 4. If the width mode bit is set, the XOR result is also stored in bit 6.
	n.resetFrequencyTimer()

	xnorResult := ^((n.lfsr & 0b01) ^ ((n.lfsr & 0b10) >> 1))
	n.lfsr = (n.lfsr >> 1) | (xnorResult & 0x1 << 14)

	if n.use7Bit {
		util.UnsetBit16(&n.lfsr, 6)
		n.lfsr |= xnorResult << 6
	}

	// The amplitude of the channel is simply the bit 0 of LFSR inverted. (Take into account envelope of-course).
	n.currentSample = byte(n.lfsr&0x01) * n.volumeEnvelope.GetVolume()
}

func (n *Noise) LengthTick() {
	if !n.lengthEnable {
		return
	}

	n.lengthCounter--
	if n.lengthCounter == 0 {
		n.Disable()
	}
}

func (n *Noise) EnvelopeTick() {
	n.volumeEnvelope.Tick()
}

func (n *Noise) Trigger() {
	if !n.volumeEnvelope.IsEnabled() {
		return
	}
	n.volumeEnvelope.Trigger()
	n.lfsr = 0
	n.enabled = true
	n.currentSample = 0
	if n.lengthCounter == 0 {
		n.lengthCounter = noiseLengthTimerMax
	}
	n.resetFrequencyTimer()
}

func (n *Noise) GetSample() byte {
	return n.currentSample
}

func (n *Noise) IsEnabled() bool {
	return n.enabled
}

func (n *Noise) Disable() {
	n.enabled = false
	n.currentSample = 0
	n.volumeEnvelope.Disable()
	n.lfsr = 0
	n.frequencyTimer = 0
	n.clockShift = 0
	n.use7Bit = false
	n.clockDivider = 0
}

// SetNRx1 sets the length timer
func (n *Noise) SetNRx1(data byte) {
	n.lengthCounter = noiseLengthTimerMax - (data & 0x3F)
}

// SetNRx2 controls the volume envelope of this channel
func (n *Noise) SetNRx2(data byte) {
	n.volumeEnvelope.Write(data)
	if !n.volumeEnvelope.IsEnabled() {
		n.Disable()
	}
}

// GetNRx2 returns the value of the NRx2 register.
func (n *Noise) GetNRx2() byte {
	return n.volumeEnvelope.Read()
}

// SetNRx3 sets the clock shift, the LFSR width and the clock divider.
func (n *Noise) SetNRx3(data byte) {
	n.clockShift = data >> 4
	n.use7Bit = util.BitIsSet8(data, 3)
	n.clockDivider = data & 0x7
}

func (n *Noise) GetNRx3() byte {
	var use7BitBit byte
	if n.use7Bit {
		use7BitBit = 0x8
	}
	return n.clockShift<<4 | use7BitBit | n.clockDivider
}

// SetNRx4 enables the length timer and triggers the channel
func (n *Noise) SetNRx4(data byte) {

	if util.BitIsSet8(data, 7) {
		n.Trigger()
	}

	n.lengthEnable = util.BitIsSet8(data, 6)
}

func (n *Noise) GetNRx4() byte {
	if n.lengthEnable {
		return 0xFF
	}
	return 0xBF
}

func (n *Noise) resetFrequencyTimer() {
	var divisor = uint(n.clockDivider) * 0x10
	if n.clockDivider == 0 {
		divisor = 0x8
	}

	n.frequencyTimer = (divisor << n.clockShift) * 16
}
