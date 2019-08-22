# mcp23xxx-periph

Microchip's MCP23xxx device driver to work with [periph](https://periph.io)

## Devices summary

MCP23xxx series are 8 or 16 bits GPIO port expanders, with SPI or I²C interface.

This driver is intended to handle devices in the following table:

Reference | I/O bits | Interface | Max bus freq | Remarks
--------- | -------- | --------- | ------------ | -------
MCP23008 | 8  | I²C | 1.7 MHz |
MCP23S08 | 8  | SPI | 10 MHz  |
MCP23009 | 8  | I²C | 3.4 MHz |
MCP23S09 | 8  | SPI | 10 MHz  |
MCP23016 | 16 | I²C | 400 kHz | Obsolete
MCP23017 | 16 | I²C | 1.7 MHz |
MCP23S17 | 16 | SPI | 10 MHz  |
MCP23018 | 16 | I²C | 3.4 MHz |
MCP23S18 | 16 | SPI | 10 MHz  |

### Datasheets

* [MCP23008/MCP23S08](http://ww1.microchip.com/downloads/en/DeviceDoc/MCP23008-MCP23S08-Data-Sheet-20001919F.pdf)
* [MCP23009/MCP23S09](http://ww1.microchip.com/downloads/en/DeviceDoc/20002121C.pdf)
* [MCP23016](http://ww1.microchip.com/downloads/en/DeviceDoc/20090C.pdf)
* [MCP23017/MCP23S17](http://ww1.microchip.com/downloads/en/DeviceDoc/20001952C.pdf)
* [MCP23018/MCP23S18](http://ww1.microchip.com/downloads/en/DeviceDoc/22103a.pdf)
