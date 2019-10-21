package mcp23xxx

import (
	"fmt"
)

type features struct {
	isSPI   bool
	regs    registers
	maxAddr uint8
}

var mcp23xxxChip = map[string]features{
	"MCP23008": {false, reg8bits, 7},
	"MCP23S08": {true, reg8bits, 3},
	"MCP23009": {false, reg8bits, 7},
	"MCP23S09": {true, reg8bits, 7},
	"MCP23017": {false, reg16bits, 7},
	"MCP23S17": {true, reg16bits, 7},
	"MCP23018": {false, reg16bits, 7},
	"MCP23S18": {true, reg16bits, 7},
}

type registers map[string]byte

var reg8bits = registers{
	"IODIR":   0x00,
	"IPOL":    0x01,
	"GPINTEN": 0x02,
	"DEFVAL":  0x03,
	"INTCON":  0x04,
	"IOCON":   0x05,
	"GPPU":    0x06,
	"INTF":    0x07,
	"INTCAP":  0x08,
	"GPIO":    0x09,
	"OLAT":    0x0A,
}

var reg16bits = registers{
	"IODIRA":   0x00,
	"IODIRB":   0x01,
	"IPOLA":    0x02,
	"IPOLB":    0x03,
	"GPINTENA": 0x04,
	"GPINTENB": 0x05,
	"DEFVALA":  0x06,
	"DEFVALB":  0x07,
	"INTCONA":  0x08,
	"INTCONB":  0x09,
	"IOCON":    0x0A,
	"GPPUA":    0x0C,
	"GPPUB":    0x0D,
	"INTFA":    0x0E,
	"INTFB":    0x0F,
	"INTCAPA":  0x10,
	"INTCAPB":  0x11,
	"GPIOA":    0x12,
	"GPIOB":    0x13,
	"OLATA":    0x14,
	"OLATB":    0x15,
}

func (r registers) addr(name string) (byte, error) {
	a, ok := r[name]
	if !ok {
		return 0, fmt.Errorf("invalid register %q", name)
	}
	return a, nil
}
