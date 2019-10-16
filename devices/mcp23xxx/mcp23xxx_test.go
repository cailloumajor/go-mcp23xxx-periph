package mcp23xxx

import (
	"fmt"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi/spitest"
)

func ifaceI2C(d *Dev) error {
	d.isI2C = true
	return nil
}

func ifaceSPI(d *Dev) error {
	d.isSPI = true
	return nil
}

func TestDevImplementsResource(t *testing.T) {
	var i interface{} = new(Dev)
	if _, ok := i.(conn.Resource); !ok {
		t.Fatalf("expected %T to implement conn.Resource", i)
	}
}

func TestNewWithSupportedModels(t *testing.T) {
	cases := []Opts{
		Opts{Model: "MCP23008", IFCfg: ifaceI2C},
		Opts{Model: "MCP23S08", IFCfg: ifaceSPI},
		Opts{Model: "MCP23009", IFCfg: ifaceI2C},
		Opts{Model: "MCP23S09", IFCfg: ifaceSPI},
		Opts{Model: "MCP23017", IFCfg: ifaceI2C},
		Opts{Model: "MCP23S17", IFCfg: ifaceSPI},
		Opts{Model: "MCP23018", IFCfg: ifaceI2C},
		Opts{Model: "MCP23S18", IFCfg: ifaceSPI},
	}
	for _, test := range cases {
		t.Run(test.Model, func(t *testing.T) {
			_, err := New(&test)
			if err != nil {
				t.Fatalf("expected no error, got %q", err)
			}
		})
	}
}

func TestHardwareAddress(t *testing.T) {
	cases := []struct {
		A0      bool
		A1      bool
		A2      bool
		expAddr uint8
	}{
		{false, false, false, 0},
		{true, false, false, 1},
		{false, false, true, 4},
		{true, true, true, 7},
	}
	for _, test := range cases {
		desc := fmt.Sprintf("{%v, %v, %v}", test.A0, test.A1, test.A2)
		o := Opts{"MCP23008", test.A0, test.A1, test.A2, ifaceI2C}
		t.Run(desc, func(t *testing.T) {
			d, _ := New(&o)
			if d.hwAddr != test.expAddr {
				t.Fatalf("expected %v, got %v", test.expAddr, d.hwAddr)
			}
		})
	}
}

func TestInterfaceConfigFunctions(t *testing.T) {
	cases := map[string]struct {
		model      string
		configFunc func(*Dev) error
		expErr     expectError
	}{
		"passing I2C()": {"MCP23008", I2C(&i2ctest.Record{}), expNoError},
		"passing SPI()": {"MCP23S08", SPI(&spitest.Record{}, 0), expNoError},
		"failing SPI()": {
			"MCP23S08", SPI(&spitest.Record{Initialized: true}, 0), expError,
		},
	}
	for desc, test := range cases {
		o := Opts{Model: test.model, IFCfg: test.configFunc}
		t.Run(desc, func(t *testing.T) {
			_, err := New(&o)
			if test.expErr != (err != nil) {
				t.Fatalf("expected %v, got \"%v\"", test.expErr, err)
			}
		})
	}
}

func TestNewWithUnsupportedModels(t *testing.T) {
	cases := []string{
		"NCP23008", "MDP23008", "MCQ23008", "MCP33008", "MCP24008", "MCP23X08",
		"MCP23007", "MCP23016", "MCP23028",
	}
	for _, test := range cases {
		t.Run(test, func(t *testing.T) {
			_, err := New(&Opts{Model: test, IFCfg: func(*Dev) error { return nil }})
			if err == nil {
				t.Fatalf("expected error, got \"%v\"", err)
			}
		})
	}
}

func TestNewWithInconsistentInterface(t *testing.T) {
	cases := map[string]Opts{
		"I²C model, SPI config": {Model: "MCP23008", IFCfg: ifaceSPI},
		"SPI model, I²C config": {Model: "MCP23S08", IFCfg: ifaceI2C},
	}
	for desc, test := range cases {
		t.Run(desc, func(t *testing.T) {
			_, err := New(&test)
			if err == nil {
				t.Fatalf("expected error, got \"%v\"", err)
			}
		})
	}
}

func TestI2CReadReg(t *testing.T) {
	cases := []struct {
		op      i2ctest.IO
		model   string
		A0      bool
		A1      bool
		A2      bool
		reg     string
		expErr  expectError
		expData byte
	}{
		// Register read success cases
		{
			i2ctest.IO{Addr: 0x20, W: []byte{0x05}, R: []byte{0x45}},
			"MCP23008", false, false, false, "IOCON", expNoError, 0x45,
		},
		{
			i2ctest.IO{Addr: 0x21, W: []byte{0x00}, R: []byte{0xff}},
			"MCP23008", true, false, false, "IODIR", expNoError, 0xff,
		},
		{
			i2ctest.IO{Addr: 0x22, W: []byte{0x09}, R: []byte{0x93}},
			"MCP23008", false, true, false, "GPIO", expNoError, 0x93,
		},
		{
			i2ctest.IO{Addr: 0x23, W: []byte{0x0a}, R: []byte{0x2b}},
			"MCP23017", true, true, false, "IOCON", expNoError, 0x2b,
		},
		{
			i2ctest.IO{Addr: 0x24, W: []byte{0x00}, R: []byte{0x8c}},
			"MCP23017", false, false, true, "IODIRA", expNoError, 0x8c,
		},
		{
			i2ctest.IO{Addr: 0x26, W: []byte{0x01}, R: []byte{0x2b}},
			"MCP23017", false, true, true, "IODIRB", expNoError, 0x2b,
		},
		{
			i2ctest.IO{Addr: 0x27, W: []byte{0x12}, R: []byte{0xd1}},
			"MCP23017", true, true, true, "GPIOA", expNoError, 0xd1,
		},
		// Register read failure cases
		{ // Bad 8 bits device register mnemonic
			i2ctest.IO{Addr: 0x24, W: []byte{0x00}, R: []byte{0x97}},
			"MCP23008", false, false, true, "IODIRA", expError, 0x00,
		},
		{ // Bad 16 bits device register mnemonic
			i2ctest.IO{Addr: 0x21, W: []byte{0x00}, R: []byte{0x74}},
			"MCP23017", true, false, false, "IODIR", expError, 0x00,
		},
		{ // Bad address
			i2ctest.IO{Addr: 0x26, W: []byte{0x12}, R: []byte{0xd1}},
			"MCP23017", true, true, true, "GPIOA", expError, 0x00,
		},
	}
	for _, test := range cases {
		desc := fmt.Sprintf(
			"%v %v,%v,%v %v", test.model, test.A0, test.A1, test.A2, test.reg,
		)
		t.Run(desc, func(t *testing.T) {
			bus := &i2ctest.Playback{Ops: []i2ctest.IO{test.op}, DontPanic: true}
			o := Opts{test.model, test.A0, test.A1, test.A2, I2C(bus)}
			d, _ := New(&o)
			rd, err := d.ReadReg(test.reg)
			if test.expErr != (err != nil) {
				t.Fatalf("expected %v, got \"%v\"", test.expErr, err)
			}
			if rd != test.expData {
				t.Fatalf("expected %#v, got %#v", test.expData, rd)
			}
		})
	}
}

func TestSPIReadReg(t *testing.T) {
	cases := []struct {
		op      conntest.IO
		model   string
		A0      bool
		A1      bool
		A2      bool
		reg     string
		expErr  expectError
		expData byte
	}{
		// Register read success cases
		{
			conntest.IO{W: []byte{0x41, 0x05, 0x00}, R: []byte{0x00, 0x00, 0x45}},
			"MCP23S08", false, false, false, "IOCON", expNoError, 0x45,
		},
		{
			conntest.IO{W: []byte{0x43, 0x00, 0x00}, R: []byte{0x00, 0x00, 0xff}},
			"MCP23S08", true, false, false, "IODIR", expNoError, 0xff,
		},
		{
			conntest.IO{W: []byte{0x45, 0x09, 0x00}, R: []byte{0x00, 0x00, 0x93}},
			"MCP23S08", false, true, false, "GPIO", expNoError, 0x93,
		},
		{
			conntest.IO{W: []byte{0x47, 0x0a, 0x00}, R: []byte{0x00, 0x00, 0x2b}},
			"MCP23S17", true, true, false, "IOCON", expNoError, 0x2b,
		},
		{
			conntest.IO{W: []byte{0x49, 0x00, 0x00}, R: []byte{0x00, 0x00, 0x8c}},
			"MCP23S17", false, false, true, "IODIRA", expNoError, 0x8c,
		},
		{
			conntest.IO{W: []byte{0x4d, 0x01, 0x00}, R: []byte{0x00, 0x00, 0x2b}},
			"MCP23S17", false, true, true, "IODIRB", expNoError, 0x2b,
		},
		{
			conntest.IO{W: []byte{0x4f, 0x12, 0x00}, R: []byte{0x00, 0x00, 0xd1}},
			"MCP23S17", true, true, true, "GPIOA", expNoError, 0xd1,
		},
		// Register read failure cases
		{ // Bad 8 bits device register mnemonic
			conntest.IO{W: []byte{0x49, 0x00, 0x00}, R: []byte{0x00, 0x00, 0x97}},
			"MCP23S08", false, false, true, "IODIRA", expError, 0x00,
		},
		{ // Bad 16 bits device register mnemonic
			conntest.IO{W: []byte{0x43, 0x00, 0x00}, R: []byte{0x00, 0x00, 0x74}},
			"MCP23S17", true, false, false, "IODIR", expError, 0x00,
		},
		{ // Bad address
			conntest.IO{W: []byte{0x4d, 0x12, 0x00}, R: []byte{0x00, 0x00, 0xd1}},
			"MCP23S17", true, true, true, "GPIOA", expError, 0x00,
		},
	}
	for _, test := range cases {
		desc := fmt.Sprintf(
			"%v %v,%v,%v %v", test.model, test.A0, test.A1, test.A2, test.reg,
		)
		t.Run(desc, func(t *testing.T) {
			port := &spitest.Playback{
				Playback: conntest.Playback{
					Ops:       []conntest.IO{test.op},
					DontPanic: true,
				},
			}
			o := Opts{test.model, test.A0, test.A1, test.A2, SPI(port, 0)}
			d, _ := New(&o)
			rd, err := d.ReadReg(test.reg)
			if test.expErr != (err != nil) {
				t.Fatalf("expected %v, got \"%v\"", test.expErr, err)
			}
			if rd != test.expData {
				t.Fatalf("expected %#v, got %#v", test.expData, rd)
			}
		})
	}
}

func TestI2CWriteReg(t *testing.T) {
	cases := []struct {
		op    i2ctest.IO
		model string
		A0    bool
		A1    bool
		A2    bool
		reg   string
		val   byte
		exp   expectError
	}{
		// Register write success cases
		{
			i2ctest.IO{Addr: 0x20, W: []byte{0x05, 0x2a}},
			"MCP23008", false, false, false, "IOCON", 0x2a, expNoError,
		},
		// Register write failure cases
		{ // Bad register mnemonic
			i2ctest.IO{Addr: 0x24, W: []byte{0x00, 0x97}},
			"MCP23008", false, false, true, "IODIRA", 0x97, expError,
		},
		{ // Bad address
			i2ctest.IO{Addr: 0x26, W: []byte{0x12, 0xd1}},
			"MCP23017", true, true, true, "GPIOA", 0xd1, expError,
		},
	}
	for _, test := range cases {
		desc := fmt.Sprintf(
			"%v %v,%v,%v %v", test.model, test.A0, test.A1, test.A2, test.reg,
		)
		t.Run(desc, func(t *testing.T) {
			bus := &i2ctest.Playback{Ops: []i2ctest.IO{test.op}, DontPanic: true}
			o := Opts{test.model, test.A0, test.A1, test.A2, I2C(bus)}
			d, _ := New(&o)
			err := d.WriteReg(test.reg, test.val)
			if test.exp != (err != nil) {
				t.Fatalf("expected %v, got \"%v\"", test.exp, err)
			}
		})
	}
}

func TestSPIWriteReg(t *testing.T) {
	cases := []struct {
		op    conntest.IO
		model string
		A0    bool
		A1    bool
		A2    bool
		reg   string
		val   byte
		exp   expectError
	}{
		// Register write success cases
		{
			conntest.IO{W: []byte{0x40, 0x05, 0x2a}},
			"MCP23S08", false, false, false, "IOCON", 0x2a, expNoError,
		},
		// Register write failure cases
		{ // Bad register mnemonic
			conntest.IO{W: []byte{0x48, 0x00, 0x97}},
			"MCP23S08", false, false, true, "IODIRA", 0x97, expError,
		},
		{ // Bad address
			conntest.IO{W: []byte{0x4c, 0x12, 0xd1}},
			"MCP23S17", true, true, true, "GPIOA", 0xd1, expError,
		},
	}
	for _, test := range cases {
		desc := fmt.Sprintf(
			"%v %v,%v,%v %v", test.model, test.A0, test.A1, test.A2, test.reg,
		)
		t.Run(desc, func(t *testing.T) {
			port := &spitest.Playback{
				Playback: conntest.Playback{
					Ops:       []conntest.IO{test.op},
					DontPanic: true,
				},
			}
			o := Opts{test.model, test.A0, test.A1, test.A2, SPI(port, 0)}
			d, _ := New(&o)
			err := d.WriteReg(test.reg, test.val)
			if test.exp != (err != nil) {
				t.Fatalf("expected %v, got \"%v\"", test.exp, err)
			}
		})
	}
}

type expectError bool

func (e expectError) String() string {
	if e {
		return "error"
	}
	return "no error"
}

const (
	expNoError expectError = false
	expError   expectError = true
)
