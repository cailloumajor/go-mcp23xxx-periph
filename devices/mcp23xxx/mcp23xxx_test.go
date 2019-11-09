package mcp23xxx

import (
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi/spitest"
)

func TestDevImplementsResource(t *testing.T) {
	var i interface{} = new(Dev)
	if _, ok := i.(conn.Resource); !ok {
		t.Fatalf("expected %T to implement conn.Resource", i)
	}
}

func TestNewNoError(t *testing.T) {
	cases := map[string]Opts{
		"I²C": {"MCP23017", 0, I2C(&i2ctest.Record{})},
		"SPI": {"MCP23S08", 0, SPI(&spitest.Record{}, 0)},
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

func TestNewError(t *testing.T) {
	cases := map[string]Opts{
		"inconsistent I²C": {
			"MCP23S09", 0, I2C(&i2ctest.Record{}),
		},
		"inconsistent SPI": {
			"MCP23009", 0, SPI(&spitest.Record{}, 0),
		},
		"SPI error": {
			"MCP23S17", 0, SPI(&spitest.Record{Initialized: true}, 0),
		},
		"unknown chip": {
			"", 0, I2C(&i2ctest.Record{}),
		},
		"hardware address too high": {
			"MCP23S08", 4, SPI(&spitest.Record{}, 0),
		},
		"missing interface configuration": {
			Model: "MCP23018", HWAddr: 0,
		},
	}
	for desc, opts := range cases {
		t.Run(desc, func(t *testing.T) {
			if d, err := New(&opts); d != nil || err == nil {
				t.Fatalf("expected (nil, error), got (%v, %v)", d, err)
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
