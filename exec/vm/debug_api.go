package vm

import "github.com/anhcraft/rice/exec/types"

// DebugPos is a paused VM location. Func is an index into Module.Functions.
type DebugPos struct {
	Func int
	IP   int
}

// SlotValue is a copy of a call-frame local at a debug pause.
type SlotValue struct {
	Name  string
	Value types.Value
}

// CallFrame is a copy of a VM call frame at a debug pause.
type CallFrame struct {
	Func  int
	IP    int
	Slots []SlotValue
}

// Debug enables single-stepping. The interpret loop only waits and publishes
// positions when built with -tags rice_debug. Call this before Run.
func (vm *VM) Debug() *VM {
	vm.debug = true
	vm.step = make(chan struct{})
	vm.curr = make(chan DebugPos)
	vm.dbgDone = make(chan struct{})
	return vm
}

// Step allows one opcode to run. After Run finishes, extra Steps are ignored.
func (vm *VM) Step() {
	step := vm.step
	done := vm.dbgDone
	if step == nil {
		return
	}
	select {
	case step <- struct{}{}:
	case <-done:
	}
}

// Position yields the next instruction after each opcode. It closes when Run returns.
// Read Stack and CallStack after receiving a position and before the next Step.
func (vm *VM) Position() <-chan DebugPos {
	return vm.curr
}

// CallStack copies the current call frames. Only call it while the VM is paused.
func (vm *VM) CallStack() []CallFrame {
	out := make([]CallFrame, len(vm.frames))
	for i, fr := range vm.frames {
		slots := make([]SlotValue, len(fr.slots))
		for j, sl := range fr.slots {
			name := ""
			if fr.fn != nil && j < len(fr.fn.SlotNames) && fr.module != nil {
				name = fr.module.ConstString(fr.fn.SlotNames[j])
			}
			slots[j] = SlotValue{Name: name, Value: sl.val}
		}
		ip := 0
		fn := 0
		if fr != nil {
			ip = fr.ip
			fn = fr.fnIndex
		}
		out[i] = CallFrame{Func: fn, IP: ip, Slots: slots}
	}
	return out
}

func (vm *VM) debugFinish() {
	if !(debug && vm.debug) {
		return
	}
	if vm.dbgDone != nil {
		close(vm.dbgDone)
	}
	if vm.curr != nil {
		close(vm.curr)
		vm.curr = nil
	}
	vm.step = nil
	vm.debug = false
}
