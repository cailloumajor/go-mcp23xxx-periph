package mcp23xxx

import (
	"reflect"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi/spitest"
)

func TestDev_implements_Resource(t *testing.T) {
	var i interface{} = new(Dev)
	if _, ok := i.(conn.Resource); !ok {
		t.Fatalf("expected %T to implement conn.Resource", i)
	}
}

func TestNew_no_error(t *testing.T) {
	cases := map[string]Opts{
		"I²C": {Model: "MCP23017", IFCfg: I2C(&i2ctest.Record{})},
		"SPI": {Model: "MCP23S08", IFCfg: SPI(&spitest.Record{}, 0)},
	}
	for desc, opts := range cases {
		t.Run(desc, func(t *testing.T) {
			var d interface{}
			d, err := New(&opts)
			got, ok := d.(*Dev)
			if !ok || err != nil {
				t.Fatalf("expected (*Dev, nil), got (%v, %v)", got, err)
			}
		})
	}
}

func TestNew_error(t *testing.T) {
	cases := map[error]Opts{
		errI2CChip: {
			Model: "MCP23S09", IFCfg: I2C(&i2ctest.Record{}),
		},
		errSPIChip: {
			Model: "MCP23009", IFCfg: SPI(&spitest.Record{}, 0),
		},
		conntest.Errorf("spitest: Connect cannot be called twice"): {
			Model: "MCP23S17",
			IFCfg: SPI(&spitest.Record{Initialized: true}, 0),
		},
		errUnknownChip: {
			IFCfg: I2C(&i2ctest.Record{}),
		},
		errHWAddrHigh: {
			Model: "MCP23S08", HWAddr: 4, IFCfg: SPI(&spitest.Record{}, 0),
		},
		errMissIntfCfg: {
			Model: "MCP23018", HWAddr: 0,
		},
	}
	for exp, opts := range cases {
		t.Run(exp.Error(), func(t *testing.T) {
			d, err := New(&opts)
			be := err.(mcp23xxxError).Unwrap()
			if d != nil {
				t.Fatalf("expected %v, got %v", nil, d)
			}
			if exp != be && !reflect.DeepEqual(exp, be) {
				t.Fatalf("expected %v, got %v", exp, be)
			}
		})
	}
}

func TestDevString(t *testing.T) {
	d := &Dev{
		c:      &conntest.Discard{},
		model:  "Model",
		hwAddr: 255,
	}
	if exp, got := "Model/discard@255", d.String(); got != exp {
		t.Fatalf("expected %q, got %q", exp, got)
	}

}
