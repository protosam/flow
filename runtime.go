package flow

import (
	_ "runtime"
	_ "unsafe"
)

//go:noescape

//go:linkname runtime_doSpin sync.runtime_doSpin
func runtime_doSpin()

//go:linkname runtime_canSpin sync.runtime_canSpin
func runtime_canSpin(i int) bool
