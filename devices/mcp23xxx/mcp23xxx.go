package mcp23xxx

import (
	"fmt"
	"regexp"
	"strings"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

// Opts holds the configuration options for the device.
type Opts struct {
	Model string           // Chip model
	A0    bool             // Logical state of A0 pin
	A1    bool             // Logical state of A1 pin
	A2    bool             // Logical state of A2 pin
	IFCfg func(*Dev) error // Interface configuration function
}

// I2C configures the device to work on I²C.
func I2C(bus i2c.Bus) func(*Dev) error {
	return func(d *Dev) error {
		c := &i2c.Dev{Bus: bus, Addr: uint16(0x20 | d.hwAddr)}
		d.isI2C = true
		d.c = c
		return nil
	}
}

// SPI configures the device to work on SPI.
func SPI(port spi.Port, f physic.Frequency) func(*Dev) error {
	return func(d *Dev) error {
		c, err := port.Connect(f, spi.Mode0, 8)
		if err != nil {
			return fmt.Errorf("SPI: %v", err)
		}
		d.isSPI = true
		d.c = c
		return nil
	}
}

// New returns a handle to a MCP23xxx I/O expander.
func New(opts *Opts) (*Dev, error) {
	var hwaddr uint8
	for i, v := range [3]bool{opts.A0, opts.A1, opts.A2} {
		if v {
			hwaddr |= 1 << uint(i)
		}
	}

	d := &Dev{hwAddr: hwaddr}

	if err := opts.IFCfg(d); err != nil {
		return nil, fmt.Errorf("mcp23xxx: %v", err)
	}

	const modelRegExp = `^MCP23([0S])(0[89]|1[78])$`
	model := strings.ToUpper(opts.Model)
	sm := regexp.MustCompile(modelRegExp).FindStringSubmatch(model)
	if len(sm) != 3 {
		return nil, fmt.Errorf("mcp23xxx: unknown chip: %q", model)
	}

	switch {
	case sm[1] == "0" && !d.isI2C:
		return nil, fmt.Errorf("mcp23xxx: chip %v must be configured with I²C", model)
	case sm[1] == "S" && !d.isSPI:
		return nil, fmt.Errorf("mcp23xxx: chip %v must be configured with SPI", model)
	}

	switch sm[2] {
	case "08", "09":
		d.regs = reg8bits
	case "17", "18":
		d.regs = reg16bits
	}

	return d, nil
}

// Dev is a handle to an initialized MCP23xxx device.
//
// It implements conn.Resource.
type Dev struct {
	hwAddr uint8
	c      conn.Conn
	isI2C  bool
	isSPI  bool
	regs   registers
}

func (d *Dev) String() string {
	return fmt.Sprintf("mcp23xxx/%v@%v", d.c, d.hwAddr)
}

// Halt halts each GPIO pin.
// FIXME implement
func (d *Dev) Halt() error {
	return fmt.Errorf("unimplemented")
}

// ReadReg reads and returns a register, given its mnemonic.
//
// This is a low-level method.
func (d *Dev) ReadReg(regName string) (byte, error) {
	addr, err := d.regs.addr(regName)
	if err != nil {
		return 0, d.wrap(err, "ReadReg")
	}
	w, r := d.makeTxData(addr, nil)
	if err := d.c.Tx(w, r); err != nil {
		return 0, d.wrap(err, "ReadReg")
	}
	return r[len(r)-1], nil
}

// WriteReg writes a register, given its mnemonic and the value to write.
//
// This is a low-level method.
func (d *Dev) WriteReg(regName string, val byte) error {
	addr, err := d.regs.addr(regName)
	if err != nil {
		return d.wrap(err, "write register")
	}
	w, r := d.makeTxData(addr, &val)
	if err := d.c.Tx(w, r); err != nil {
		return d.wrap(err, "write register")
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
func (d *Dev) makeTxData(regAddr byte, write *byte) (w, r []byte) {
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
