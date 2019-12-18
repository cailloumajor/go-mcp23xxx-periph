package mcp23xxx

import "fmt"

type baseError string

func (e baseError) Error() string {
	return string(e)
}

const (
	errI2CChip       = baseError("bad I²C chip")
	errSPIChip       = baseError("bad SPI chip")
	errUnknownChip   = baseError("unknown chip")
	errHWAddrHigh    = baseError("hardware address too high")
	errMissIntfCfg   = baseError("missing interface configuration function")
	errUnknownINTCfg = baseError("unknown INT pin configuration")
)

type mcp23xxxError struct {
	orig fmt.Stringer
	op   string
	err  error
}

func (e mcp23xxxError) Error() string {
	return fmt.Sprintf("%v %v: %v", e.orig, e.op, e.err)
}

func (e mcp23xxxError) Unwrap() error {
	return e.err
}
