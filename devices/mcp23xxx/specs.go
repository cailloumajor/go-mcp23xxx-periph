package mcp23xxx

type features struct {
	isSPI    bool
	is16bits bool
	maxAddr  uint8
	conf     byte
}

var mcp23xxxChip = map[string]features{
	"MCP23008": {false, false, 7, cSEQOP},
	"MCP23S08": {true, false, 3, cHAEN | cSEQOP},
	"MCP23009": {false, false, 7, cSEQOP},
	"MCP23S09": {true, false, 7, cSEQOP},
	"MCP23017": {false, true, 7, cSEQOP | cBANK},
	"MCP23S17": {true, true, 7, cHAEN | cSEQOP | cBANK},
	"MCP23018": {false, true, 7, cSEQOP | cBANK},
	"MCP23S18": {true, true, 7, cSEQOP | cBANK},
}

const (
	// Register addresses.
	rIODIR   byte = iota // 0x00
	rIPOL                // 0x00
	rGPINTEN             // 0x01
	rDEFVAL              // 0x02
	rINTCON              // 0x03
	rIOCON               // 0x04
	rGPPU                // 0x05
	rINTF                // 0x06
	rINTCAP              // 0x07
	rGPIO                // 0x08
	rOLAT                // 0x09
)

const (
	// Configuration register bits.
	cINTCC  byte = 1 << iota // Bit 0
	cINTPOL                  // Bit 1
	cODR                     // Bit 2
	cHAEN                    // Bit 3
	cDISSLW                  // Bit 4
	cSEQOP                   // Bit 5
	cMIRROR                  // Bit 6
	cBANK                    // Bit 7
)
