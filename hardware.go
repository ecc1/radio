package radio

import (
	"fmt"
	"log"
	"time"

	"github.com/ecc1/gpio"
	"github.com/ecc1/spi"
)

// HardwareFlavor is the interface satisfied by a particular SPI device.
// It specifies how to open the device and how to encode common I/O operations.
type HardwareFlavor interface {
	SPIDevice() string
	Speed() int
	CustomCS() int
	InterruptPin() int
	ReadSingleAddress(byte) byte
	ReadBurstAddress(byte) byte
	WriteSingleAddress(byte) byte
	WriteBurstAddress(byte) byte
}

// Hardware represents an SPI radio device.
type Hardware struct {
	device    *spi.Device
	flavor    HardwareFlavor
	err       error
	interrupt gpio.InterruptPin
}

// Device returns the radio's SPI device pathname.
func (h *Hardware) Device() string {
	return h.flavor.SPIDevice()
}

// Error returns the error state of the radio device.
func (h *Hardware) Error() error {
	return h.err
}

// SetError sets the error state of the radio device.
func (h *Hardware) SetError(err error) {
	h.err = err
}

// AwaitInterrupt waits with the given timeout for a receive interrupt.
func (h *Hardware) AwaitInterrupt(timeout time.Duration) {
	h.err = h.interrupt.Wait(timeout)
}

// ReadInterrupt returns the state of the receive interrupt.
func (h *Hardware) ReadInterrupt() bool {
	b, err := h.interrupt.Read()
	h.err = err
	return b
}

// Open opens the SPI radio module described by the given flavor.
func Open(flavor HardwareFlavor) *Hardware {
	h := &Hardware{flavor: flavor}
	h.device, h.err = spi.Open(flavor.SPIDevice(), flavor.Speed(), flavor.CustomCS())
	if h.Error() != nil {
		return h
	}
	h.err = h.device.SetMaxSpeed(flavor.Speed())
	if h.Error() != nil {
		h.Close()
		return h
	}
	h.interrupt, h.err = gpio.Interrupt(flavor.InterruptPin(), false, "rising")
	if h.Error() != nil {
		h.Close()
		return h
	}
	return h
}

// Close closes the radio device.
func (h *Hardware) Close() {
	h.err = h.device.Close()
}

// ReadRegister reads the given address on the radio device.
func (h *Hardware) ReadRegister(addr byte) byte {
	if h.Error() != nil {
		return 0
	}
	buf := []byte{h.flavor.ReadSingleAddress(addr), 0}
	h.err = h.device.Transfer(buf)
	return buf[1]
}

// ReadBurst reads a burst of n bytes from given address on the radio device.
func (h *Hardware) ReadBurst(addr byte, n int) []byte {
	if h.Error() != nil {
		return nil
	}
	buf := make([]byte, n+1)
	buf[0] = h.flavor.ReadBurstAddress(addr)
	h.err = h.device.Transfer(buf)
	return buf[1:]
}

// WriteRegister writes the given value to the given address on the radio device.
func (h *Hardware) WriteRegister(addr byte, value byte) {
	h.err = h.device.Write([]byte{h.flavor.WriteSingleAddress(addr), value})
}

// WriteBurst writes data in burst mode to the given address on the radio device.
func (h *Hardware) WriteBurst(addr byte, data []byte) {
	h.err = h.device.Write(append([]byte{h.flavor.WriteBurstAddress(addr)}, data...))
}

// WriteEach writes each address-value pairs in data to the radio device.
func (h *Hardware) WriteEach(data []byte) {
	n := len(data)
	if n%2 != 0 {
		log.Panicf("odd data length (%d)", n)
	}
	for i := 0; i < n; i += 2 {
		h.WriteRegister(data[i], data[i+1])
	}
}

// SPIDevice returns the radio's SPI device.
func (h *Hardware) SPIDevice() *spi.Device {
	return h.device
}

// HardwareVersionError indicates a hardware version mismatch.
type HardwareVersionError struct {
	Actual   uint16
	Expected uint16
}

func (e HardwareVersionError) Error() string {
	return fmt.Sprintf("unexpected hardware version %04X (should be %04X)", e.Actual, e.Expected)
}
