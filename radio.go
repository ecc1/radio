package radio

import (
	"fmt"
	"time"
)

// Interface is the interface satisfied by a radio device.
type Interface interface {
	Init(frequency uint32)
	Reset()
	Close()

	Frequency() uint32
	SetFrequency(uint32)

	Send([]byte)
	Receive(time.Duration) ([]byte, int)
	SendAndReceive([]byte, time.Duration) ([]byte, int)

	State() string

	Error() error
	SetError(error)

	Name() string
	Device() string
}

// MegaHertz converts a frequency in Hertz into a string denoting
// that frequency in MegaHertz, with 3 decimal places (kiloHertz).
func MegaHertz(freq uint32) string {
	m := freq / 1000000
	k := (freq % 1000000) / 1000
	return fmt.Sprintf("%3d.%03d", m, k)
}
