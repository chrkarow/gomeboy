package main

import (
	"gameboy-emulator/internal/cycle/emulation"
	"github.com/ebitengine/oto/v3"
	"io"
)

const playerBufferSize = 4096
const readAhead = 512

type SoundDriver struct {
	ctx     *oto.Context
	pl      *oto.Player
	core    *emulation.Core
	stopped bool
	muted   bool
}

func NewSoundDriver(ctx *oto.Context, core *emulation.Core) *SoundDriver {
	return &SoundDriver{
		ctx:  ctx,
		core: core,
	}
}

func (d *SoundDriver) Run() {
	d.stopped = false
	d.pl = d.ctx.NewPlayer(d)
	d.pl.SetBufferSize(playerBufferSize)
	d.pl.SetVolume(0)
	d.pl.Play()
	d.pl.SetVolume(1)
}

func (d *SoundDriver) TogglePause() {
	if d.pl == nil {
		return
	}

	if d.IsPaused() {
		d.pl.Play()
	} else {
		d.pl.Pause()
	}
}

func (d *SoundDriver) IsPaused() bool {
	return d.pl != nil && !d.pl.IsPlaying()
}

func (d *SoundDriver) Stop() {
	d.stopped = true
	d.pl.SetVolume(0)
	d.pl.Pause()
	err := d.pl.Close()
	if err != nil {
		panic(err)
	}
	d.pl = nil
	d.core.SaveGame()
	d.core.Reset()
}

func (d *SoundDriver) GetCore() *emulation.Core {
	return d.core
}

func (d *SoundDriver) Read(buffer []byte) (int, error) {
	for i := 0; i < readAhead; {
		if d.stopped {
			return 0, io.EOF
		}

		left, right, play := d.core.Tick()

		if play {
			buffer[i] = left
			buffer[i+1] = right
			i += 2
		}
	}
	return readAhead, nil
}

func (d *SoundDriver) ToggleMute() {
	if d.pl == nil {
		return
	}

	if d.pl.Volume() == 1 {
		d.pl.SetVolume(0)
	} else {
		d.pl.SetVolume(1)
	}
}
