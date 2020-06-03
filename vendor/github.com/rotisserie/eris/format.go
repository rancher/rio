package eris

import (
	"fmt"
)

// Format defines an error output format to be used with the default formatter.
type Format struct {
	WithTrace bool   // Flag that enables stack trace output.
	Msg       string // Separator between error messages and stack frame data.
	TBeg      string // Separator at the beginning of each stack frame.
	TSep      string // Separator between elements of each stack frame.
	Sep       string // Separator between each error in the chain.
}

// NewDefaultFormat conveniently returns a basic format for the default string formatter.
func NewDefaultFormat(withTrace bool) Format {
	stringFmt := Format{
		WithTrace: withTrace,
		Sep:       ": ",
	}
	if withTrace {
		stringFmt.Msg = "\n"
		stringFmt.TBeg = "\t"
		stringFmt.TSep = ": "
		stringFmt.Sep = "\n"
	}
	return stringFmt
}

// UnpackedError represents complete information about an error.
//
// This type can be used for custom error logging and parsing. Use `eris.Unpack` to build an UnpackedError
// from any error type. The ErrChain and ErrRoot fields correspond to `wrapError` and `rootError` types,
// respectively. If any other error type is unpacked, it will appear in the ExternalErr field.
type UnpackedError struct {
	ErrChain    *[]ErrLink
	ErrRoot     *ErrRoot
	ExternalErr string
}

// Unpack returns UnpackedError type for a given golang error type.
func Unpack(err error) UnpackedError {
	e := UnpackedError{}
	switch err.(type) {
	case nil:
		return UnpackedError{}
	case *rootError:
		e = unpackRootErr(err.(*rootError))
	case *wrapError:
		chain := []ErrLink{}
		e = unpackWrapErr(&chain, err.(*wrapError))
	default:
		e.ExternalErr = err.Error()
	}
	return e
}

// ToString returns a default formatted string for a given eris error.
func (upErr *UnpackedError) ToString(format Format) string {
	var str string
	if upErr.ErrChain != nil {
		for _, eLink := range *upErr.ErrChain {
			str += eLink.formatStr(format)
		}
	}
	str += upErr.ErrRoot.formatStr(format)
	if upErr.ExternalErr != "" {
		str += fmt.Sprint(upErr.ExternalErr)
	}
	return str
}

// ToJSON returns a JSON formatted map for a given eris error.
func (upErr *UnpackedError) ToJSON(format Format) map[string]interface{} {
	if upErr == nil {
		return nil
	}
	jsonMap := make(map[string]interface{})
	if fmtRootErr := upErr.ErrRoot.formatJSON(format); fmtRootErr != nil {
		jsonMap["error root"] = fmtRootErr
	}
	if upErr.ErrChain != nil {
		var wrapArr []map[string]interface{}
		for _, eLink := range *upErr.ErrChain {
			wrapMap := eLink.formatJSON(format)
			wrapArr = append(wrapArr, wrapMap)
		}
		jsonMap["error chain"] = wrapArr
	}
	if upErr.ExternalErr != "" {
		jsonMap["external error"] = fmt.Sprint(upErr.ExternalErr)
	}
	return jsonMap
}

func unpackRootErr(err *rootError) UnpackedError {
	return UnpackedError{
		ErrRoot: &ErrRoot{
			Msg:   err.msg,
			Stack: err.stack.get(),
		},
	}
}

func unpackWrapErr(chain *[]ErrLink, err *wrapError) UnpackedError {
	link := ErrLink{}
	link.Frame = *err.frame.get()
	link.Msg = err.msg
	*chain = append(*chain, link)

	e := UnpackedError{}
	e.ErrChain = chain

	nextErr := err.Unwrap()
	switch nextErr.(type) {
	case nil:
		return e
	case *rootError:
		uErr := unpackRootErr(nextErr.(*rootError))
		e.ErrRoot = uErr.ErrRoot
	case *wrapError:
		e = unpackWrapErr(chain, nextErr.(*wrapError))
	default:
		e.ExternalErr = err.Error()
	}
	return e
}

type ErrRoot struct {
	Msg   string
	Stack []StackFrame
}

func (err *ErrRoot) formatStr(format Format) string {
	if err == nil {
		return ""
	}
	str := err.Msg
	str += format.Msg
	if format.WithTrace {
		stackArr := formatStackFrames(err.Stack, format.TSep)
		for _, frame := range stackArr {
			str += format.TBeg
			str += frame
			str += format.Sep
		}
	}
	return str
}

func (err *ErrRoot) formatJSON(format Format) map[string]interface{} {
	if err == nil {
		return nil
	}
	rootMap := make(map[string]interface{})
	rootMap["message"] = fmt.Sprint(err.Msg)
	if format.WithTrace {
		rootMap["stack"] = formatStackFrames(err.Stack, format.TSep)
	}
	return rootMap
}

type ErrLink struct {
	Msg   string
	Frame StackFrame
}

func (eLink *ErrLink) formatStr(format Format) string {
	var str string
	str += eLink.Msg
	str += format.Msg
	if format.WithTrace {
		str += format.TBeg
		str += eLink.Frame.formatFrame(format.TSep)
	}
	str += format.Sep
	return str
}

func (eLink *ErrLink) formatJSON(format Format) map[string]interface{} {
	wrapMap := make(map[string]interface{})
	wrapMap["message"] = fmt.Sprint(eLink.Msg)
	if format.WithTrace {
		wrapMap["stack"] = eLink.Frame.formatFrame(format.TSep)
	}
	return wrapMap
}

func formatStackFrames(s []StackFrame, sep string) []string {
	var str []string
	for _, f := range s {
		str = append(str, f.formatFrame(sep))
	}
	return str
}
