# GOmeBoy

A GameBoy emulator written in Go.

This project actually contains two models:

- `step/` and `internal/step`contain an emulator model which follows a stepped approach in which the CPU is returning
  how many steps it has needed to execute an particular instruction.
- `cyle/` and `internal/cycle` contain the more detailed and advanced model based on single clock cycles (T-Cycles) and
  an attempt to recreate the actual way the GameBoy's PPU is drawing pixels to the screen.

## Main Sources

- PanDocs: https://gbdev.io/pandocs
- OpCode table: https://izik1.github.io/gbops/
- OpCode reference: https://rgbds.gbdev.io/docs/v0.9.1/gbz80.7
- Gameboy Emulator Step-by-step: http://www.codeslinger.co.uk/pages/projects/gameboy/banking.html

- Emulator-Sourcen in C++: https://github.com/CTurt/Cinoop/blob/master/source/cpu.c#L429
- Gameboy Emulator Developer Guide: https://hacktix.github.io/GBEDG/

## Other stuff

- DMG-01 - How to emulate a GameBoy: https://rylev.github.io/DMG-01/public/book/introduction.html
- Sound Emulation : https://nightshade256.github.io/2021/03/27/gb-sound-emulation.html
- Another emulator (in JS): https://imrannazar.com/series/gameboy-emulation-in-javascript/memory
- https://gbdev.gg8.se/wiki/articles/Gameboy_Bootstrap_ROM
- https://b13rg.icecdn.tech/Gameboy-Bank-Switching/
- https://raphaelstaebler.medium.com/memory-and-memory-mapped-i-o-of-the-gameboy-part-3-of-a-series-37025b40d89b

