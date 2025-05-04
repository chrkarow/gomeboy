package interrupts

const (
	VBlank InterruptType = 1 << iota
	LcdStat
	Timer
	Serial
	Joypad
)

type (
	InterruptSink interface {
		RequestInterrupt(t InterruptType)
	}
	InterruptType byte
	Interrupts    struct {
		master bool
		enable byte
		flags  byte
	}
)

var interruptPriority = []InterruptType{
	VBlank,
	LcdStat,
	Timer,
	Serial,
	Joypad,
}

func New() *Interrupts {
	i := &Interrupts{}
	i.Reset()
	return i
}

func (i *Interrupts) Reset() {
	i.master = false
	i.enable = 0xE0
	i.flags = 0xE0
}

func (i *Interrupts) SetEnable(value byte) {
	i.enable = value | 0xE0 // Sets upper 3 bits to 1 (= unused)
}

func (i *Interrupts) SetFlags(value byte) {
	i.flags = value | 0xE0 // though the upper bits are unused, it is a full 8-byte register -> no masking
}

func (i *Interrupts) RequestInterrupt(t InterruptType) {
	i.flags |= byte(t)
}

func (i *Interrupts) GetEnable() byte {
	return i.enable
}

func (i *Interrupts) GetFlags() byte {
	return i.flags
}

func (i *Interrupts) SetMasterEnable(value bool) {
	i.master = value
}

func (i *Interrupts) MasterEnabled() bool {
	return i.master
}

func (i *Interrupts) HandleInterrupt(interruptCallback func(InterruptType)) {
	if !i.master || !i.InterruptsPending() {
		return
	}

	fire := i.enable & i.flags

	for _, t := range interruptPriority {
		if fire&byte(t) == byte(t) {
			i.master = false
			i.flags &= ^byte(t)
			interruptCallback(t)
			return // only handle interrupts one by one
		}
	}
}

func (i *Interrupts) InterruptsPending() bool {
	return i.enable&i.flags != 0xE0
}

func (i *Interrupts) MustHandleInterrupt() bool {
	return i.master && i.InterruptsPending()
}
