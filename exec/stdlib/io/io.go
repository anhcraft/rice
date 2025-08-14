package io

import (
	"context"
	"fmt"
	"io"
	"rice/exec/ctxkey"
	"rice/exec/fun"
	"rice/exec/stdlib"
	"rice/exec/types"
	"rice/exec/types/values"
)

var Functions = fun.FunctionPackage{
	"print":    {stdlib.Define(Print)},
	"printf":   {stdlib.Define(Printf)},
	"println":  {stdlib.Define(Println)},
	"printlnf": {stdlib.Define(Printlnf)},
}

// Print writes the given values to standard output, separated by spaces, and returns nil.
func Print(ctx context.Context, values ...any) (types.Value, error) {
	out := ctx.Value(ctxkey.LoggingOutput).(io.Writer)
	_, err := fmt.Fprint(out, values...)
	return nil, err
}

// Println writes the given values to standard output, separated by spaces, and returns nil.
func Println(ctx context.Context, values ...any) (types.Value, error) {
	out := ctx.Value(ctxkey.LoggingOutput).(io.Writer)
	_, err := fmt.Fprintln(out, values...)
	return nil, err
}

// Printf formats according to a format specifier and writes to standard output.
func Printf(ctx context.Context, format values.String, values ...any) (types.Value, error) {
	out := ctx.Value(ctxkey.LoggingOutput).(io.Writer)
	_, err := fmt.Fprintf(out, string(format), values...)
	return nil, err
}

// Printlnf formats according to a format specifier and writes to standard output with an ending newline.
func Printlnf(ctx context.Context, format values.String, values ...any) (types.Value, error) {
	out := ctx.Value(ctxkey.LoggingOutput).(io.Writer)
	_, err := fmt.Fprintf(out, string(format)+"\n", values...)
	return nil, err
}
