package mcp23xxx

import (
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi/spitest"
)

func TestDevImplementsResource(t *testing.T) {
	var i interface{} = new(Dev)
	if _, ok := i.(conn.Resource); !ok {
		t.Fatalf("expected %T to implement conn.Resource", i)
	}
}

func TestNew(t *testing.T) {
	cases := map[string]struct {
		o   Opts
		exp expectError
	}{
		"alright I²C": {
			Opts{"MCP23017", 0, I2C(&i2ctest.Record{})}, expNoError,
		},
		"alright SPI": {
			Opts{"MCP23S08", 0, SPI(&spitest.Record{}, 0)}, expNoError,
		},
		"inconsistent I²C": {
			Opts{"MCP23S09", 0, I2C(&i2ctest.Record{})}, expError,
		},
		"inconsistent SPI": {
			Opts{"MCP23009", 0, SPI(&spitest.Record{}, 0)}, expError,
		},
		"SPI error": {
			Opts{"MCP23S17", 0, SPI(&spitest.Record{Initialized: true}, 0)}, expError,
		},
		"unknown chip": {
			Opts{"", 0, I2C(&i2ctest.Record{})}, expError,
		},
		"hardware address too high": {
			Opts{"MCP23S08", 4, SPI(&spitest.Record{}, 0)}, expError,
		},
		"missing interface configuration": {
			Opts{Model: "MCP23018", HWAddr: 0}, expError,
		},
	}
	for desc, test := range cases {
		t.Run(desc, func(t *testing.T) {
			_, err := New(&test.o)
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
