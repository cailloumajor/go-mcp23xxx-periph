package mcp23xxx

import "fmt"

type mcp23xxxError struct {
	orig fmt.Stringer
	op   string
	err  error
}

func (e *mcp23xxxError) Error() string {
	return fmt.Sprintf("%v %v: %v", e.orig, e.op, e.err)
}

func (e *mcp23xxxError) Unwrap() error {
	return e.err
}
