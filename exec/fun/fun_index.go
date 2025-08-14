package fun

import "rice/exec/types/values"

// FunctionImpl is a list of overloading definitions (expected to be under the same name)
// e.g. [substr0, substr1]
// where substr0: func(string, off)
//
//	substr1: func(string, off, end)
type FunctionImpl = []*FunctionDef

// FunctionPackage is a package of functions under the same logical group
// e.g. String: {"substr": [substr0, substr1], "slice": [slice0, slice1]}
type FunctionPackage = map[values.Identifier]FunctionImpl
