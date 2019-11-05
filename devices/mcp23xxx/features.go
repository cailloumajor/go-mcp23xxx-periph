package mcp23xxx

type features struct {
	isSPI    bool
	is16bits bool
	maxAddr  uint8
}

var mcp23xxxChip = map[string]features{
	"MCP23008": {false, false, 7},
	"MCP23S08": {true, false, 3},
	"MCP23009": {false, false, 7},
	"MCP23S09": {true, false, 7},
	"MCP23017": {false, true, 7},
	"MCP23S17": {true, true, 7},
	"MCP23018": {false, true, 7},
	"MCP23S18": {true, true, 7},
}

type register byte

const (
	// Register addresses.
	rIODIR      register = 0x00
	rIPOL       register = 0x01
	rGPINTEN    register = 0x02
	rDEFVAL     register = 0x03
	rINTCON     register = 0x04
	rIOCON      register = 0x05
	rGPPU       register = 0x06
	rINTF       register = 0x07
	rINTCAP     register = 0x08
	rGPIO       register = 0x09
	rOLAT       register = 0x0A
	rIOCONBANK0 byte     = 0x0B
)

func (r register) addr(bankB bool) byte {
	if bankB {
		return byte(r) | 0x10
	}
	return byte(r)
}
