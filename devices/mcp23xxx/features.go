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

type regAddr byte

const (
	// Register addresses.
	rIODIR      regAddr = 0x00
	rIPOL       regAddr = 0x01
	rGPINTEN    regAddr = 0x02
	rDEFVAL     regAddr = 0x03
	rINTCON     regAddr = 0x04
	rIOCON      regAddr = 0x05
	rGPPU       regAddr = 0x06
	rINTF       regAddr = 0x07
	rINTCAP     regAddr = 0x08
	rGPIO       regAddr = 0x09
	rOLAT       regAddr = 0x0A
	rIOCONBANK0 regAddr = 0x0B
)
