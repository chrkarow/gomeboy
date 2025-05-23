package apu

type (
	FrameSequencer struct {
		channels []WaveGenerator

		step    byte
		tick    uint
		enabled bool
	}
)

func NewFrameSequencer(channels ...WaveGenerator) *FrameSequencer {
	return &FrameSequencer{
		channels: channels,
	}
}

func (f *FrameSequencer) Tick() {
	if f.tick++; f.tick == 8192 { // 512Hz
		f.tick = 0
		f.sequencerStep()
	}

	for _, c := range f.channels {
		c.Tick()
	}
}

func (f *FrameSequencer) SetEnabled(value bool) {
	if f.enabled && !value {
		for _, c := range f.channels {
			c.Disable()
		}
	}
	f.enabled = value
}

func (f *FrameSequencer) sequencerStep() {
	if f.enabled {
		for _, c := range f.channels {
			if f.step%2 == 0 {
				c.LengthTick()
			}

			if s, ok := c.(Sweeper); ok && (f.step == 2 || f.step == 6) {
				s.SweepTick()
			}

			if e, ok := c.(Enveloper); ok && f.step == 7 {
				e.EnvelopeTick()
			}
		}
	}

	f.step = (f.step + 1) % 8
}
