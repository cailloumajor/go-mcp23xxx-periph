package mcp23xxx

import (
	"fmt"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

// Opts holds the configuration options for the device.
type Opts struct {
	Model  string           // Chip model
	HWAddr uint8            // Hardware address (refer to datasheet)
	IFCfg  func(*Dev) error // Interface configuration function
}

// I2C configures the device to work on I²C.
func I2C(bus i2c.Bus) func(*Dev) error {
	return func(d *Dev) error {
		if d.isSPI {
			return fmt.Errorf("inconsistent chip model and interface")
		}
		c := &i2c.Dev{Bus: bus, Addr: uint16(0x20 | d.hwAddr)}
		d.c = c
		return nil
	}
}

// SPI configures the device to work on SPI.
func SPI(port spi.Port, f physic.Frequency) func(*Dev) error {
	return func(d *Dev) error {
		if !d.isSPI {
			return fmt.Errorf("inconsistent chip model and interface")
		}
		c, err := port.Connect(f, spi.Mode0, 8)
		if err != nil {
			return fmt.Errorf("SPI: %v", err)
		}
		d.c = c
		return nil
	}
}

// New returns a handle to a MCP23xxx I/O expander.
func New(opts *Opts) (*Dev, error) {
	d := &Dev{hwAddr: opts.HWAddr}

	f, ok := mcp23xxxChip[opts.Model]
	if !ok {
		return nil, fmt.Errorf("mcp23xxx: unknown chip: %q", opts.Model)
	}

	d.model, d.isSPI, d.is16bits = opts.Model, f.isSPI, f.is16bits

	if opts.HWAddr > f.maxAddr {
		return nil, fmt.Errorf(
			"mcp23xxx: maximum hardware address for %v is %v",
			d.model, f.maxAddr,
		)
	}

	if opts.IFCfg == nil {
		return nil, fmt.Errorf(
			"mcp23xxx: missing interface configuration function",
		)
	}
	if err := opts.IFCfg(d); err != nil {
		return nil, fmt.Errorf("mcp23xxx: %v", err)
	}

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
func (d *Dev) readReg(ra regAddr) (byte, error) {
	w, r := d.makeTxData(ra, nil)
	if err := d.c.Tx(w, r); err != nil {
		return 0, d.wrap(err, "readReg")
	}
	return r[len(r)-1], nil
}

// writeReg writes a register, given its address and the value to write.
func (d *Dev) writeReg(ra regAddr, val byte) error {
	w, r := d.makeTxData(ra, &val)
	if err := d.c.Tx(w, r); err != nil {
		return d.wrap(err, "writeReg")
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
func (d *Dev) makeTxData(ra regAddr, write *byte) (w, r []byte) {
	regAddr := byte(ra)
	if d.is16bits {
		regAddr |= 0x10
	}
	w = append(w, regAddr)
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

func (d *Dev) wrap(err error, ctx string) error {
	return fmt.Errorf("%v: %v: %v", d, ctx, err)
}
