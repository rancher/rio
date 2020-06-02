package eris

import (
	"fmt"
	"runtime"
	"strings"
)

// StackFrame stores a frame's runtime information in a human readable format.
type StackFrame struct {
	Name string
	File string
	Line int
}

func (f *StackFrame) formatFrame(sep string) string {
	return fmt.Sprintf("%v%v%v%v%v", f.Name, sep, f.File, sep, f.Line)
}

// caller returns a single stack frame. the argument skip is the number of stack frames
// to ascend, with 0 identifying the caller of Caller.
func caller(skip int) *frame {
	pc, _, _, _ := runtime.Caller(skip)
	var f frame = frame(pc)
	return &f
}

// callers returns a stack trace. the argument skip is the number of stack frames to skip
// before recording in pc, with 0 identifying the frame for Callers itself and 1 identifying
// the caller of Callers.
func callers(skip int) *stack {
	const depth = 64
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// frame is a single program counter of a stack frame.
type frame uintptr

func (f frame) pc() uintptr {
	return uintptr(f) - 1
}

func (f frame) get() *StackFrame {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return &StackFrame{
			Name: "unknown",
			File: "unknown",
		}
	}

	name := fn.Name()
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	file, line := fn.FileLine(f.pc())

	return &StackFrame{
		Name: name,
		File: file,
		Line: line,
	}
}

// stack is an array of program counters.
type stack []uintptr

func (s *stack) get() []StackFrame {
	var sFrames []StackFrame
	for _, f := range *s {
		frame := frame(f)
		sFrame := frame.get()
		sFrames = append(sFrames, *sFrame)
	}
	return sFrames
}
