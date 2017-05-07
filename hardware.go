package radio

import (
	"fmt"
	"log"
	"time"

	"github.com/ecc1/gpio"
	"github.com/ecc1/spi"
)

type HardwareFlavor interface {
	Name() string
	SPIDevice() string
	Speed() int
	CustomCS() int
	InterruptPin() int
	ReadSingleAddress(byte) byte
	ReadBurstAddress(byte) byte
	WriteSingleAddress(byte) byte
	WriteBurstAddress(byte) byte
}

type Hardware struct {
	device    *spi.Device
	flavor    HardwareFlavor
	err       error
	interrupt gpio.InputPin
}

func (h *Hardware) Name() string {
	return h.flavor.Name()
}

func (h *Hardware) Device() string {
	return h.flavor.SPIDevice()
}

func (h *Hardware) Error() error {
	return h.err
}

func (h *Hardware) SetError(err error) {
	h.err = err
}

func (h *Hardware) AwaitInterrupt(timeout time.Duration) {
	h.err = h.interrupt.Wait(timeout)
}

func (h *Hardware) ReadInterrupt() bool {
	b, err := h.interrupt.Read()
	h.err = err
	return b
}

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
	h.interrupt, h.err = gpio.Input(flavor.InterruptPin(), "rising", false)
	if h.Error() != nil {
		h.Close()
		return h
	}
	return h
}

func (h *Hardware) Close() {
	h.device.Close()
}

func (h *Hardware) ReadRegister(addr byte) byte {
	if h.Error() != nil {
		return 0
	}
	buf := []byte{h.flavor.ReadSingleAddress(addr), 0}
	h.err = h.device.Transfer(buf)
	return buf[1]
}

func (h *Hardware) ReadBurst(addr byte, n int) []byte {
	if h.Error() != nil {
		return nil
	}
	buf := make([]byte, n+1)
	buf[0] = h.flavor.ReadBurstAddress(addr)
	h.err = h.device.Transfer(buf)
	return buf[1:]
}

func (h *Hardware) WriteRegister(addr byte, value byte) {
	h.err = h.device.Write([]byte{h.flavor.WriteSingleAddress(addr), value})
}

func (h *Hardware) WriteBurst(addr byte, data []byte) {
	h.err = h.device.Write(append([]byte{h.flavor.WriteBurstAddress(addr)}, data...))
}

func (h *Hardware) WriteEach(data []byte) {
	n := len(data)
	if n%2 != 0 {
		log.Panicf("odd data length (%d)", n)
	}
	for i := 0; i < n; i += 2 {
		h.WriteRegister(data[i], data[i+1])
	}
}

func (h *Hardware) SPIDevice() *spi.Device {
	return h.device
}

type HardwareVersionError struct {
	Actual   uint16
	Expected uint16
}

func (e HardwareVersionError) Error() string {
	return fmt.Sprintf("unexpected hardware version %04X (should be %04X)", e.Actual, e.Expected)
}
