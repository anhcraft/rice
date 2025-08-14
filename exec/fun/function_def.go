package fun

import (
	"errors"
	"reflect"
	"rice/exec/types"
	"strings"
)

type FunctionDef struct {
	handler    reflect.Value
	args       [][]ArgType
	variadic   bool
	contextual bool
}

func ScanFunction(fun any) (*FunctionDef, error) {
	tp := reflect.TypeOf(fun)
	if tp.Kind() != reflect.Func {
		return nil, errors.New("fun must be a function")
	}

	var args [][]ArgType
	contextual := false

	if tp.NumIn() > 0 && tp.In(0).Implements(typeOfContext) {
		contextual = true

		args = make([][]ArgType, tp.NumIn()-1)

		for i := 0; i < tp.NumIn()-1; i++ {
			argType, err := getArgType(tp.In(i + 1))
			if err != nil {
				return nil, err
			}

			args[i] = []ArgType{argType}
		}
	} else {
		args = make([][]ArgType, tp.NumIn())

		for i := 0; i < tp.NumIn(); i++ {
			argType, err := getArgType(tp.In(i))
			if err != nil {
				return nil, err
			}

			args[i] = []ArgType{argType}
		}
	}

	return &FunctionDef{
		handler:    reflect.ValueOf(fun),
		args:       args,
		variadic:   tp.IsVariadic(),
		contextual: contextual,
	}, nil
}

func (f *FunctionDef) DefineArg(i int, t ...ArgType) {
	if len(t) > 0 {
		f.args[i] = t
	}
}

func (f *FunctionDef) Arg(i int) []ArgType {
	return f.args[i]
}

func (f *FunctionDef) SizeOfArgs() int {
	return len(f.args)
}

func (f *FunctionDef) String() string {
	sb := strings.Builder{}
	sb.WriteString("(")
	for i, arg := range f.args {
		if i > 0 {
			sb.WriteString(",")
		}
		if f.variadic && i == len(f.args)-1 {
			sb.WriteString("...")
		}

		for j, tp := range arg {
			if j > 0 {
				sb.WriteString("|")
			}
			sb.WriteString(tp.String())
		}
	}
	sb.WriteString(")")
	return sb.String()
}

func (f *FunctionDef) ReadableString() any {
	sb := strings.Builder{}
	sb.WriteString("(")
	for i, arg := range f.args {
		if i > 0 {
			sb.WriteString(",")
		}

		if f.variadic && i == len(f.args)-1 {
			sb.WriteString("...")
		}

		for j, tp := range arg {
			if j > 0 {
				sb.WriteString("|")
			}
			if tp.IsAny() {
				sb.WriteString("any")
			} else {
				sb.WriteString(types.Type(tp.Type()).String())
				if tp.Dim() > 0 {
					sb.WriteString(strings.Repeat("[]", int(tp.Dim())))
				}
			}
		}
	}
	sb.WriteString(")")
	return sb.String()
}
