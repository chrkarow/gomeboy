package interrupts

import log "go.uber.org/zap"

type Handlers interface {
	HandleVblankInterrupt()
	HandleLcdStatInterrupt()
	HandleTimerInterrupt()
	HandleSerialInterrupt()
	HandleJoypadInterrupt()
}

type loggingHandlers struct{}

func (h *loggingHandlers) HandleVblankInterrupt() {
	log.L().Debug("VBlank interrupt occured!")
}

func (h *loggingHandlers) HandleLcdStatInterrupt() {
	log.L().Debug("LCD Stat interrupt occured!")
}

func (h *loggingHandlers) HandleTimerInterrupt() {
	log.L().Debug("Timer interrupt occured!")
}

func (h *loggingHandlers) HandleSerialInterrupt() {
	log.L().Debug("Serial interrupt occured!")
}

func (h *loggingHandlers) HandleJoypadInterrupt() {
	log.L().Debug("Joypad interrupt occured!")
}
