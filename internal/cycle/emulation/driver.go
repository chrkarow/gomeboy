package emulation

import "time"

type (
	// A Driver is responsible for driving the emulation by ticking its core repeatedly.
	Driver interface {
		Run()
		TogglePause()
		IsPaused() bool
		Stop()
		GetCore() *Core
	}

	BasicDriver struct {
		core    *Core
		paused  bool
		stopped bool
		turbo   bool
	}
)

func NewBasicEmulator(core *Core) *BasicDriver {
	return &BasicDriver{
		core: core,
	}
}

func (d *BasicDriver) Run() {
	go func() {

		// save cartridge RAM when emulator loop ends
		defer d.core.SaveGame()

		var count int
		for !d.stopped {

			if d.paused {
				time.Sleep(time.Second)
				continue
			}

			if !d.turbo && count == 10000 {
				count = 0
				time.Sleep(1600 * time.Microsecond)
			}

			d.core.Tick()

			count++
		}

		d.core.Reset()
		d.stopped = false
		d.paused = false
	}()
}

func (d *BasicDriver) ToggleTurbo() {
	d.turbo = !d.turbo
}

func (d *BasicDriver) TogglePause() {
	d.paused = !d.paused
}

func (d *BasicDriver) IsPaused() bool {
	return d.paused
}

func (d *BasicDriver) Stop() {
	d.stopped = true
}

func (d *BasicDriver) GetCore() *Core {
	return d.core
}
