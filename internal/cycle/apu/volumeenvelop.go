package apu

import "gameboy-emulator/internal/util"

type VolumeEnvelope struct {
	initial   byte // NRx2 bits 7-4
	increase  bool // NRx2 bit 3
	sweepPace byte // NRx2 bits 2-0

	volume byte // Current calculated volume.

	enabled     bool
	periodTimer byte

	ticks byte
}

func NewVolumeEnvelope() *VolumeEnvelope {
	return &VolumeEnvelope{}
}

func (e *VolumeEnvelope) Tick() {
	if !e.enabled {
		return
	}

	if e.periodTimer--; e.periodTimer > 0 {
		return
	}

	e.periodTimer = e.sweepPace

	if (e.volume == 0xF && e.increase) || (e.volume == 0x0 && !e.increase) || e.sweepPace == 0 {
		return
	}

	if e.increase {
		e.volume++
	} else {
		e.volume--
	}
}

func (e *VolumeEnvelope) Trigger() {
	if !e.enabled {
		return
	}
	e.periodTimer = e.sweepPace
	e.volume = e.initial
}

func (e *VolumeEnvelope) Write(data byte) {
	e.initial = data >> 4
	e.increase = util.BitIsSet8(data, 3)
	e.sweepPace = data & 0x7
	e.enabled = e.initial != 0 && !e.increase
	if !e.enabled {
		e.volume = 0
		e.periodTimer = 0
	}
}

func (e *VolumeEnvelope) Read() byte {
	var increaseBit byte
	if e.increase {
		increaseBit = 0x8
	}
	return e.initial<<4 | increaseBit | e.sweepPace
}

func (e *VolumeEnvelope) GetVolume() byte {
	return e.volume
}

func (e *VolumeEnvelope) Disable() {
	e.initial = 0
	e.increase = false
	e.sweepPace = 0
	e.volume = 0
	e.enabled = false
	e.periodTimer = 0
	e.ticks = 0
}

func (e *VolumeEnvelope) IsEnabled() bool {
	return e.enabled
}
