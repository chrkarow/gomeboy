package interrupts

const (
	VBlank byte = 1 << iota
	LcdStat
	Timer
	Serial
	Joypad
)

type Interrupts struct {
	master bool
	enable byte
	flags  byte

	handlers []Handlers
}

func New() *Interrupts {
	i := &Interrupts{}
	i.Reset()
	i.handlers = append(i.handlers, &loggingHandlers{})
	return i
}

func (i *Interrupts) Reset() {
	i.master = false
	i.enable = 0x0
	i.flags = 0xE1
}

func (i *Interrupts) RegisterHandlers(handlers Handlers) {
	i.handlers = append(i.handlers, handlers)
}

func (i *Interrupts) SetEnable(value byte) {
	i.enable = value
}

func (i *Interrupts) GetEnable() byte {
	return i.enable
}

func (i *Interrupts) SetFlags(value byte) {
	i.flags = value
}

func (i *Interrupts) GetFlags() byte {
	return i.flags
}

func (i *Interrupts) RequestInterrupt(flag byte) {
	i.flags |= flag
}

func (i *Interrupts) SetMasterEnable(value bool) {
	i.master = value
}

func (i *Interrupts) MasterEnabled() bool {
	return i.master
}

func (i *Interrupts) HandleInterrupt() {
	if !i.master || !i.InterruptsPending() {
		return
	}

	fire := i.enable & i.flags

	if fire&VBlank == VBlank {
		i.master = false
		i.flags &= ^VBlank
		i.notifyHandlers(VBlank)
	}

	if fire&LcdStat == LcdStat {
		i.master = false
		i.flags &= ^LcdStat
		i.notifyHandlers(LcdStat)
	}

	if fire&Timer == Timer {
		i.master = false
		i.flags &= ^Timer
		i.notifyHandlers(Timer)
	}

	if fire&Serial == Serial {
		i.master = false
		i.flags &= ^Serial
		i.notifyHandlers(Serial)
	}

	if fire&Joypad == Joypad {
		i.master = false
		i.flags &= ^Joypad
		i.notifyHandlers(Joypad)
	}
}

func (i *Interrupts) InterruptsPending() bool {
	return i.enable&i.flags != 0x00
}

func (i *Interrupts) notifyHandlers(interruptType byte) {
	for _, handler := range i.handlers {
		switch interruptType {
		case VBlank:
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
