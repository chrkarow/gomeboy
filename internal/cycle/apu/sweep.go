package apu

import "gameboy-emulator/internal/util"

type (
	Sweeper interface {
		Enveloper
		SweepTick()
	}

	SweepableSquareWave struct {
		pace           byte
		subtract       bool
		individualStep byte // or sweep shift

		enabled      bool
		shadowPeriod uint
		sweepTimer   byte

		squareWave *SquareWave
	}
)

func NewSweepableSquareWave(sq *SquareWave) *SweepableSquareWave {
	return &SweepableSquareWave{
		squareWave: sq,
	}
}

func (s *SweepableSquareWave) Tick() {
	s.squareWave.Tick()
}

func (s *SweepableSquareWave) LengthTick() {
	s.squareWave.LengthTick()
}

func (s *SweepableSquareWave) EnvelopeTick() {
	s.squareWave.EnvelopeTick()
}

func (s *SweepableSquareWave) SweepTick() {
	if s.sweepTimer--; s.sweepTimer > 0 {
		return
	}

	if s.pace == 0 {
		s.sweepTimer = 0x8
	} else {
		s.sweepTimer = s.pace
	}

	if s.enabled && s.pace != 0 {
		newPeriod, overflow := s.calculateNewPeriod()

		if overflow {
			s.squareWave.Disable()
		}

		if s.individualStep != 0 {
			s.squareWave.period = newPeriod
			s.shadowPeriod = newPeriod

			if _, o := s.calculateNewPeriod(); o {
				s.squareWave.Disable()
			}
		}
	}
}

func (s *SweepableSquareWave) GetSample() byte {
	return s.squareWave.GetSample()
}

func (s *SweepableSquareWave) Trigger() {
	s.shadowPeriod = s.squareWave.period
	if s.pace == 0 {
		s.sweepTimer = 0x8
	} else {
		s.sweepTimer = s.pace
	}
	s.enabled = s.pace != 0x0 || s.individualStep != 0x0

	if s.individualStep != 0x0 {
		if _, overflow := s.calculateNewPeriod(); overflow {
			s.squareWave.Disable()
		}
	}
}

func (s *SweepableSquareWave) IsEnabled() bool {
	return s.squareWave.IsEnabled()
}

func (s *SweepableSquareWave) Disable() {
	s.enabled = false
	s.pace = 0
	s.subtract = false
	s.individualStep = 0
	s.shadowPeriod = 0
	s.sweepTimer = 0
	s.squareWave.Disable()
}

func (s *SweepableSquareWave) calculateNewPeriod() (newPeriod uint, overflow bool) {
	periodAdj := s.shadowPeriod >> s.individualStep

	if s.subtract {
		newPeriod = s.shadowPeriod - periodAdj
	} else {
		newPeriod = s.shadowPeriod + periodAdj
	}

	if newPeriod > 0x7FF {
		overflow = true
	}

	return
}

func (s *SweepableSquareWave) SetNRx0(data byte) {
	s.pace = data >> 4
	s.subtract = util.BitIsSet8(data, 3)
	s.individualStep = data & 0x7
}

func (s *SweepableSquareWave) GetNRx0() byte {
	var subtractBit byte
	if s.subtract {
		subtractBit = 0x8
	}
	return 1<<7 | s.pace<<4 | subtractBit | s.individualStep
}

func (s *SweepableSquareWave) SetNRx1(data byte) {
	s.squareWave.SetNRx1(data)
}

func (s *SweepableSquareWave) GetNRx1() byte {
	return s.squareWave.GetNRx1()
}

func (s *SweepableSquareWave) SetNRx2(data byte) {
	s.squareWave.SetNRx2(data)
}

func (s *SweepableSquareWave) GetNRx2() byte {
	return s.squareWave.GetNRx2()
}

func (s *SweepableSquareWave) SetNRx3(data byte) {
	s.squareWave.SetNRx3(data)
}

func (s *SweepableSquareWave) SetNRx4(data byte) {
	s.squareWave.SetNRx4(data)
	if util.BitIsSet8(data, 7) {
		s.Trigger()
	}
}

func (s *SweepableSquareWave) GetNRx4() byte {
	return s.squareWave.GetNRx4()
}
