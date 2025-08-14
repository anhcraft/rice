package stdlib

import (
	"github.com/anhcraft/rice/exec/fun"
	"reflect"
)

// Define scans the given function value; panic if any error occurs
func Define(f any) *fun.FunctionDef {
	def, err := fun.ScanFunction(f)
	if err != nil {
		panic(err)
	}
	return def
}

// DefineAndMap scans the given function value; panic if any error occurs
// then execute the callback to modify the function definition
func DefineAndMap(f any, callback func(def *fun.FunctionDef)) *fun.FunctionDef {
	def, err := fun.ScanFunction(f)
	if err != nil {
		panic(err)
	}
	callback(def)
	return def
}

// GetDimension finds the dimension size; minimum is zero
func GetDimension(t reflect.Type) int {
	dim := 0
	for t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		dim++
		t = t.Elem()
	}
	return dim
}

// GetBaseType gets the base type of vector
func GetBaseType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}
	return t
}
