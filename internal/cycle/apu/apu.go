package apu

type APU struct {
	nr [60]byte
}

func New() *APU {
	a := &APU{}
	a.Reset()
	return a
}

func (a *APU) Reset() {
	a.nr[10] = 0x80 // bit 7 unused (= 1)
	a.nr[11] = 0xBF
	a.nr[13] = 0xFF
	a.nr[14] = 0x38 // bit 3, 4 and 5 unused (= 1)
	a.nr[21] = 0x3F
	a.nr[23] = 0xFF
	a.nr[24] = 0xBF // bit 3, 4 and 5 unused (= 1)
	a.nr[30] = 0x7F // bit 0-6 unused (= 1)
	a.nr[31] = 0xFF
	a.nr[32] = 0x9F // bit 0-4 and 7 unused (= 1)
	a.nr[33] = 0xFF
	a.nr[34] = 0xBF // bit 3, 4 and 5 unused (= 1)
	a.nr[41] = 0xFF
	a.nr[44] = 0xBF // bit 0-5 unused (= 1)

	a.nr[52] = 0xF1 // bits 4,5 and 6 unused (= 1)
}

func (a *APU) ReadNR(nrNumber byte) byte {
	return a.nr[nrNumber]
}

func (a *APU) WriteNR(nrNumber byte, data byte) {
	switch nrNumber {
	case 10:
		a.nr[nrNumber] = data | 0x80
	case 13, 23, 33:
		return

	case 14, 24, 34:
		a.nr[nrNumber] = data | 0x38
	case 30:
		a.nr[nrNumber] = data | 0x7F
	case 32:
		a.nr[nrNumber] = data | 0x9F
	case 41:
		return
	case 44, 11, 16:
		a.nr[nrNumber] = data | 0x3F
	case 52:
		a.nr[nrNumber] = data | 0x71
	default:
		a.nr[nrNumber] = data
	}
}
