package interrupts

const (
	Vblank  byte = 1 << 0
	LcdStat byte = 1 << 1
	Timer   byte = 1 << 2
	Serial  byte = 1 << 3
	Joypad  byte = 1 << 4
)

type Interrupts struct {
	master bool
	enable byte
	flags  byte

	handlers []Handlers
}

func New() *Interrupts {
	i := &Interrupts{
		flags: 0xE1,
	}
	i.handlers = append(i.handlers, &loggingHandlers{})
	return i
}

func (i *Interrupts) Reset() {
	i.master = false
	i.enable = 0x0
}

func (interrupts *Interrupts) RegisterHandlers(handlers Handlers) {
	interrupts.handlers = append(interrupts.handlers, handlers)
}

func (interrupts *Interrupts) SetEnable(value byte) {
	interrupts.enable = value
}

func (interrupts *Interrupts) SetFlags(value byte) {
	interrupts.flags = value
}

func (interrupts *Interrupts) RequestInterrupt(flag byte) {
	interrupts.flags |= flag
}

func (interrupts *Interrupts) GetEnable() byte {
	return interrupts.enable
}

func (interrupts *Interrupts) GetFlags() byte {
	return interrupts.flags
}

func (interrupts *Interrupts) SetMasterEnable(value bool) {
	interrupts.master = value
}

func (interrupts *Interrupts) IsMasterEnabled() bool {
	return interrupts.master
}

func (interrupts *Interrupts) HandleInterrupt() {
	if !interrupts.master || interrupts.enable == 0x00 || interrupts.flags == 0x00 {
		return
	}

	fire := interrupts.enable & interrupts.flags

	if fire&Vblank == Vblank {
		interrupts.master = false
		interrupts.flags &= ^Vblank
		interrupts.notifyHandlers(Vblank)
	}

	if fire&LcdStat == LcdStat {
		interrupts.master = false
		interrupts.flags &= ^LcdStat
		interrupts.notifyHandlers(LcdStat)
	}

	if fire&Timer == Timer {
		interrupts.master = false
		interrupts.flags &= ^Timer
		interrupts.notifyHandlers(Timer)
	}

	if fire&Serial == Serial {
		interrupts.master = false
		interrupts.flags &= ^Serial
		interrupts.notifyHandlers(Serial)
	}

	if fire&Joypad == Joypad {
		interrupts.master = false
		interrupts.flags &= ^Joypad
		interrupts.notifyHandlers(Joypad)
	}
}

func (interrupts *Interrupts) notifyHandlers(interruptType byte) {
	for _, handler := range interrupts.handlers {
		switch interruptType {
		case Vblank:
			handler.HandleVblankInterrupt()
		case LcdStat:
			handler.HandleLcdStatInterrupt()
		case Timer:
			handler.HandleTimerInterrupt()
		case Serial:
			handler.HandleSerialInterrupt()
		case Joypad:
			handler.HandleJoypadInterrupt()
		}
	}
}
