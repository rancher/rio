// Package eris provides a better way to handle, trace, and log errors in Go.
//
// Types of errors
//
// This package is concerned with only three different types of errors: root
// errors, wrap errors, and external errors. Root and wrap errors are defined
// types in this package and all other error types are external or third-party
// errors.
//
// Root errors are created via eris.New and eris.Errorf. Generally, it's a
// good idea to maintain a set of root errors that are then wrapped with
// additional context whenever an error of that type occurs. Wrap errors
// represent a stack of errors that have been wrapped with additional context.
// Unwrapping these errors via eris.Unwrap will return the next error in the
// stack until a root error is reached. eris.Cause will also retrieve the root
// error.
//
// When external error types are wrapped with additional context, a root error
// is first created from the original error. This creates a stack trace for the
// error and allows it to function with the rest of the `eris` package.
//
// Wrapping errors with additional context
//
// eris.Wrap adds context to an error while preserving the type of the
// original error. This method behaves differently for each error type. For
// root errors, the stack trace is reset to the current callers which ensures
// traces are correct when using global/sentinel error values. Wrapped error
// types are simply wrapped with the new context. For external types (i.e.
// something other than root or wrap errors), a new root error is created for
// the original error and then it's wrapped with the additional context.
//
// 		_, err := db.Get(id)
// 		if err != nil {
//			// return the error with some useful context
// 			return eris.Wrapf(err, "error getting resource '%v'", id)
// 		}
//
// Inspecting error types
//
// The eris package provides a few ways to inspect and compare error types.
// eris.Is returns true if a particular error appears anywhere in the error
// chain, and eris.Cause returns the root cause of the error. Currently,
// eris.Is works simply by comparing error messages with each other. If an
// error contains a particular error message anywhere in its chain (e.g. "not
// found"), it's defined to be that error type (i.e. eris.Is will return
// true).
//
// 		NotFound := eris.New("not found")
// 		_, err := db.Get(id)
//		// check if the resource was not found
// 		if eris.Is(err, NotFound) || eris.Cause(err) == NotFound {
//			// return the error with some useful context
//			return eris.Wrapf(err, "error getting resource '%v'", id)
//		}
//
// Stack traces
//
// Errors created with this package contain stack traces that are managed
// automatically even when wrapping global errors or errors from other
// libraries. Stack traces are currently mandatory when creating and wrapping
// errors but optional when printing or logging errors. Printing an error with
// or without the stack trace is simple:
//
//		_, err := db.Get(id)
//		if err != nil {
//			return eris.Wrapf(err, "error getting resource '%v'", id)
//		}
//		fmt.Printf("%v", err) // print without the stack trace
//		fmt.Printf("%+v", err) // print with the stack trace
//
// For an error that has been wrapped once, the output will look something
// like this:
//
//		# output without the stack trace
//		error getting resource 'example-id': not found
//
//		# output with the stack trace
//		error getting resource 'example-id'
//			api.GetResource: /path/to/file/api.go: 30
//		not found
//			api.GetResource: /path/to/file/api.go: 30
//			db.Get: /path/to/file/db.go: 99
//			runtime.goexit: /path/to/go/src/libexec/src/runtime/asm_amd64.s: 1337
//
// The first layer of the full error output shows a message ("error getting
// resource 'example-id'") and a single stack frame. The next layer shows the
// root error ("not found") and the full stack trace.
//
// Logging errors with more control
//
// While eris supports logging errors with Go's fmt package, it's often
// advantageous to use the provided string and JSON formatters instead. These
// methods provide much more control over the error output and should work
// seamlessly with whatever logging package you choose.
//
//    var fields log.Fields
//    unpackedErr := eris.Unpack(err)
//    fields["method"] = "api.GetResource"
//    fields["error"] = unpackedErr.ToJSON(eris.NewDefaultFormat(true))
//    logger.WithFields(fields).Errorf("method completed with error (%v)", err)
//
// When using a JSON logger, the output should look something like this:
//
//    {
//      "method":"api.GetResource",
//      "error":{
//        "error chain":[
//          {
//            "message":"error getting resource 'example-id'",
//            "stack":"api.GetResource: /path/to/file/api.go: 30"
//          }
//        ],
//        "error root":{
//          "message":"not found",
//          "stack":[
//            "api.GetResource: /path/to/file/api.go: 30",
//            "db.Get: /path/to/file/db.go: 99",
//            "runtime.goexit: /path/to/go/src/runtime/asm_amd64.s: 1337"
//          ]
//        }
//      }
//    }
//
package eris

import (
	"fmt"
	"io"
	"reflect"
)

// New creates a new root error with a static message.
func New(msg string) error {
	return &rootError{
		msg:   msg,
		stack: callers(3),
	}
}

// Errorf creates a new root error with a formatted message.
func Errorf(format string, args ...interface{}) error {
	return &rootError{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(3),
	}
}

// Wrap adds additional context to all error types while maintaining the type of the original error.
//
// This method behaves differently for each error type. For root errors, the stack trace is reset to the current
// callers which ensures traces are correct when using global/sentinel error values. Wrapped error types are simply
// wrapped with the new context. For external types (i.e. something other than root or wrap errors), a new root
// error is created for the original error and then it's wrapped with the additional context.
func Wrap(err error, msg string) error {
	return wrap(err, msg)
}

// Wrapf adds additional context to all error types while maintaining the type of the original error.
//
// This is a convenience method for wrapping errors with formatted messages and is otherwise the same as Wrap.
func Wrapf(err error, format string, args ...interface{}) error {
	return wrap(err, fmt.Sprintf(format, args...))
}

func wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *rootError:
		e.stack = callers(4)
	case *wrapError:
	default:
		err = &rootError{
			msg:   e.Error(),
			stack: callers(4),
		}
	}

	return &wrapError{
		msg:   msg,
		err:   err,
		frame: caller(3),
	}
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type contains an Unwrap method
// returning error. Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if it implements a method
// Is(error) bool such that Is(target) returns true.
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflect.TypeOf(target).Comparable()
	for {
		if isComparable && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

// Cause returns the root cause of the error, which is defined as the first error in the chain. The original
// error is returned if it does not implement `Unwrap() error` and nil is returned if the error is nil.
func Cause(err error) error {
	for {
		uerr := Unwrap(err)
		if uerr == nil {
			return err
		}
		err = uerr
	}
}

type rootError struct {
	msg   string
	stack *stack
}

func (e *rootError) Error() string {
	return fmt.Sprint(e)
}

func (e *rootError) Format(s fmt.State, verb rune) {
	printError(e, s, verb)
}

func (e *rootError) Is(target error) bool {
	if err, ok := target.(*rootError); ok {
		return e.msg == err.msg
	}
	return e.msg == target.Error()
}

type wrapError struct {
	msg   string
	err   error
	frame *frame
}

func (e *wrapError) Error() string {
	return fmt.Sprint(e)
}

func (e *wrapError) Format(s fmt.State, verb rune) {
	printError(e, s, verb)
}

func (e *wrapError) Is(target error) bool {
	if err, ok := target.(*wrapError); ok {
		return e.msg == err.msg
	}
	return e.msg == target.Error()
}

func (e *wrapError) Unwrap() error {
	return e.err
}

func printError(err error, s fmt.State, verb rune) {
	var withTrace bool
	switch verb {
	case 'v':
		if s.Flag('+') {
			withTrace = true
		}
	}
	format := NewDefaultFormat(withTrace)
	uErr := Unpack(err)
	str := uErr.ToString(format)
	_, _ = io.WriteString(s, str)
}
