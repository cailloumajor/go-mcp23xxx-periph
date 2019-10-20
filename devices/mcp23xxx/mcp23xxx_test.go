package mcp23xxx

import (
	"fmt"
	"testing"

	"periph.io/x/periph/conn"
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
