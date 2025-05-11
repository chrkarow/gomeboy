package main

import (
	"gameboy-emulator/internal/cycle/emulation"
	"github.com/ebitengine/oto/v3"
	"io"
)

const playerBufferSize = 4096
const readAhead = 1024

type SoundDriver struct {
	ctx     *oto.Context
	pl      *oto.Player
	core    *emulation.Core
	stopped bool
}

func NewSoundDriver(ctx *oto.Context, core *emulation.Core) *SoundDriver {
	return &SoundDriver{
		ctx:  ctx,
		core: core,
	}
}

func (e *SoundDriver) Run() {
	e.stopped = false
	e.pl = e.ctx.NewPlayer(e)
	e.pl.SetBufferSize(playerBufferSize)
	e.pl.SetVolume(0)
	e.pl.Play()
	e.pl.SetVolume(1)
}

func (e *SoundDriver) TogglePause() {
	if e.IsPaused() {
		e.pl.Play()
	} else {
		e.pl.Pause()
	}
}

func (e *SoundDriver) IsPaused() bool {
	return e.pl != nil && !e.pl.IsPlaying()
}

func (e *SoundDriver) Stop() {
	e.stopped = true
	e.pl.SetVolume(0)
	e.pl.Pause()
	err := e.pl.Close()
	if err != nil {
		panic(err)
	}
	e.pl = nil
	e.core.SaveGame()
	e.core.Reset()
}

func (e *SoundDriver) GetCore() *emulation.Core {
	return e.core
}

func (e *SoundDriver) Read(buffer []byte) (int, error) {
	for i := 0; i < readAhead; {
		if e.stopped {
			return 0, io.EOF
		}

		left, right, play := e.core.Tick()
		if play {
			buffer[i] = left
			buffer[i+1] = right
			i += 2
		}
	}
	return readAhead, nil
}
