package radio

import (
	"fmt"
	"time"
)

// Counters represents amounts of data sent and received.
type Counters struct {
	Sent     int
	Received int
}

// Statistics counts the number of bytes and packets.
type Statistics struct {
	Bytes   Counters
	Packets Counters
}

// Interface is the interface satisfied by a radio device.
type Interface interface {
	Init(frequency uint32)
	Reset()
	Close()

	Frequency() uint32
	SetFrequency(uint32)

	Send([]byte)
	Receive(time.Duration) ([]byte, int)

	State() string
	Statistics() Statistics

	Error() error
	SetError(error)

	Hardware() *Hardware
}

// MegaHertz converts a frequency in Hertz into a string denoting
// that frequency in MegaHertz, with 3 decimal places (kiloHertz).
func MegaHertz(freq uint32) string {
	m := freq / 1000000
	k := (freq % 1000000) / 1000
	return fmt.Sprintf("%3d.%03d", m, k)
}
