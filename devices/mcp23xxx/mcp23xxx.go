package mcp23xxx

import (
	"fmt"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

// Opts holds the configuration options for the device.
type Opts struct {
	// Chip model.
	Model string
	// Hardware address (refer to datasheet).
	HWAddr uint8
	// Interface configuration function.
	IFCfg
	// GPIO pin for capturing interrupts. If given, it must be already configured.
	IRQPin gpio.PinIn
	// INT pin configuration.
	INTPinFunc
}

// IFCfg represents a function to configure the communication interface.
type IFCfg func(*Dev) error

// I2C configures the device to work on I²C.
func I2C(bus i2c.Bus) IFCfg {
	return func(d *Dev) error {
		if d.isSPI {
			return errI2CChip
		}
		c := &i2c.Dev{Bus: bus, Addr: uint16(0x20 | d.hwAddr)}
		d.c = c
		return nil
	}
}

// SPI configures the device to work on SPI.
func SPI(port spi.Port, f physic.Frequency) IFCfg {
	return func(d *Dev) error {
		if !d.isSPI {
			return errSPIChip
		}
		c, err := port.Connect(f, spi.Mode0, 8)
		if err != nil {
			return err
		}
		d.c = c
		return nil
	}
}

// INTPinFunc represents the configuration of INT pin.
// Refer to datasheet for possible modes.
type INTPinFunc int

// Possible INT pin configurations.
const (
	// Active driver, active-low (default).
	INTActiveLow INTPinFunc = iota
	// Active driver, active-high.
	INTActiveHigh
	// Open-drain.
	INTOpenDrain
)

// New returns a handle to a MCP23xxx I/O expander.
func New(opts *Opts) (*Dev, error) {
	d := &Dev{
		model:  opts.Model,
		hwAddr: opts.HWAddr,
	}

	bad := func(err error) (*Dev, error) { return nil, d.wrap("New()", err) }

	f, ok := mcp23xxxChip[opts.Model]
	if !ok {
		return bad(errUnknownChip)
	}
	d.isSPI, d.is16bits = f.isSPI, f.is16bits
	if opts.HWAddr > f.maxAddr {
		return bad(errHWAddrHigh)
	}

	if opts.IFCfg == nil {
		return bad(errMissIntfCfg)
	}
	if err := opts.IFCfg(d); err != nil {
		return bad(err)
	}

	iocon := f.conf
	if opts.IRQPin != nil {
		d.irqPIN = opts.IRQPin
		switch opts.INTPinFunc {
		case INTActiveLow:
		case INTActiveHigh:
			iocon |= cINTPOL
		case INTOpenDrain:
			iocon |= cODR
		default:
			return bad(errUnknownINTCfg)
		}
	}
	d.writeReg(rIOCON, iocon)

	return d, nil
}

// Dev is a handle to an initialized MCP23xxx device.
//
// It implements conn.Resource.
type Dev struct {
	c        conn.Conn
	model    string
	hwAddr   uint8
	isSPI    bool
	is16bits bool
	irqPIN   gpio.PinIn

	sync.Mutex
}

// String returns a human readable identifier representing this resource in a
// descriptive way for the user (implements conn.Resource).
func (d *Dev) String() string {
	return fmt.Sprintf("%v/%v@%v", d.model, d.c, d.hwAddr)
}

// Halt halts each GPIO pin (implements conn.Resource).
func (d *Dev) Halt() error { // FIXME implement
	return fmt.Errorf("unimplemented")
}

// readReg reads and returns a register, given its address.
func (d *Dev) readReg(ra byte) (byte, error) {
	d.Lock()
	defer d.Unlock()
	return d.readRegUnderLock(ra)
}

// writeReg writes a register, given its address and the value to write.
func (d *Dev) writeReg(ra, val byte) error {
	d.Lock()
	defer d.Unlock()
	return d.writeRegUnderLock(ra, val)
}

func (d *Dev) updateReg(ra, mask byte, set bool) error {
	d.Lock()
	defer d.Unlock()
	rv, err := d.readRegUnderLock(ra)
	if err != nil {
		return err
	}
	if set {
		rv = rv | mask
	} else {
		rv = rv &^ mask
	}
	return d.writeRegUnderLock(ra, rv)
}

// readRegUnderLock reads and returns a register, given its address.
// It is intended to be called under mutex lock.
func (d *Dev) readRegUnderLock(ra byte) (byte, error) {
	w, r := d.makeTxData(ra, nil)
	if err := d.c.Tx(w, r); err != nil {
		return 0, d.wrap("read register", err)
	}
	return r[len(r)-1], nil
}

// writeRegUnderLock writes a register, given its address and the value to write.
// It is intended to be called under mutex lock.
func (d *Dev) writeRegUnderLock(ra, val byte) error {
	w, r := d.makeTxData(ra, &val)
	if err := d.c.Tx(w, r); err != nil {
		return d.wrap("write register", err)
	}
	return nil
}

// makeTxData returns write and read buffers to be used by conn.Tx().
//
// Data flow follows table below.
//
//                READ             WRITE
// -------- ---------------- ----------------
//  I²C Tx        0xRR           0xRR 0xWW
//      Rx        0xDD
// -------- ---------------- ----------------
//  SPI Tx   0xCC 0xRR 0X00   0xCC 0xRR 0xWW
//      Rx   0x00 0x00 0xDD
//
// 0xRR: register address
// 0xCC: control byte (SPI)
// 0xDD: data read
// 0xWW: data to write
func (d *Dev) makeTxData(ra byte, write *byte) (w, r []byte) {
	w = append(w, ra)
	if d.isSPI {
		ctrlByte := (0x20 | d.hwAddr) << 1
		w = append([]byte{ctrlByte}, w...) // Prepend control byte
		if write == nil {
			w[0] |= 0x01
			w = append(w, 0x00)
		}
	}
	if write != nil {
		w = append(w, *write)
	} else {
		r = make([]byte, len(w))
	}
	return
}

func (d *Dev) wrap(op string, err error) error {
	return mcp23xxxError{d, op, err}
}
