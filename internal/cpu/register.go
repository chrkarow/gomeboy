package cpu

import "fmt"

type SubRegister struct {
	value byte
}

type Register struct {
	Hi SubRegister
	Lo SubRegister
}

func newRegister() *Register {
	return &Register{SubRegister{}, SubRegister{}}
}

// GetValue combines the bytes from its two sub-registers to a 16bit value.
func (r *Register) GetValue() uint16 {
	return (uint16(r.Hi.value) << 8) + uint16(r.Lo.value)
}

func (r *Register) SetValue(val uint16) {
	r.Hi.value = byte(val >> 8)
	r.Lo.value = byte(val)
}

func (r *Register) Increment() {
	value := r.GetValue()
	value++
	r.SetValue(value)
}

func (r *Register) Decrement() {
	value := r.GetValue()
	value--
	r.SetValue(value)
}

func (r *Register) String() string {
	return fmt.Sprintf("0x%04X", r.GetValue())
}

func (sr *SubRegister) GetValue() byte {
	return sr.value
}

func (sr *SubRegister) SetValue(val byte) {
	sr.value = val
}

func (sr *SubRegister) Increment() {
	value := sr.GetValue()
	value++
	sr.SetValue(value)
}

func (sr *SubRegister) Decrement() {
	value := sr.GetValue()
	value--
	sr.SetValue(value)
}

func (sr *SubRegister) String() string {
	return fmt.Sprintf("0x%02X", sr.GetValue())
}
