package errors

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Kind is simply a string, but it allows New to function the way it does, and limits what can be
// passed as the kind of an error to things defined as actual error kinds.
type Kind string

// Error is a general-purpose error type, providing much more contextual information and utility
// when compared to the built-in error interface.
type Error struct {
	// Kind can be used as a sort of pseudo-type that check on. It's a useful mechanism for avoiding
	// "sentinel" errors, or for checking an error's type. Kind is defined as a string so that error
	// kinds can be defined in other packages.
	Kind Kind

	// Message is a human-readable, user-friendly string. Unlike caller, Message is really intended
	// to be user-facing, i.e. safe to send to the front-end.
	Message string

	// Cause is the previous error. The error that triggered this error. If it is nil, then the root
	// cause is this Error instance. If Cause is not nil, but also not of type Error, then the root
	// cause is the error in Cause.
	Cause error

	// Fields is a general-purpose map for storing key/value information. Useful for providing
	// additional structured information in logs.
	Fields map[string]interface{}

	// caller is the function that was called when this error occurred. Useful for identifying where
	// an error occurred, or providing information to developers (i.e. this should not be revealed
	// or used in responses / sent to the front-end). This may be something as simple as the method
	// name being called, or perhaps include more information to do with parameters.
	caller string

	// Stack location information.
	file string
	line int
}

// Error satisfies the standard library's error interface. It returns a message that should be
// useful as part of logs, as that's where this method will likely be used most, including the
// caller, and the message, for the whole stack.
func (e *Error) Error() string {
	return e.format(false)
}

// Format allows this error to be formatted differently, depending on the needs of the developer.
// The different formatting options made available are:
//
// %v:  Standard formatting: shows callers, and shows messages, for the whole stack.
// %+v: Verbose formatting: shows callers, and shows messages, for the whole stack, with file and
//      line, information, across multiple lines.
func (e *Error) Format(s fmt.State, c rune) {
	if c == 'v' && s.Flag('+') {
		io.WriteString(s, e.format(true))
		return
	}

	io.WriteString(s, e.format(false))
}

// WithFields appends a set of key/value pairs to the error's field list.
func (e *Error) WithFields(kvs ...interface{}) *Error {
	kvc := len(kvs)

	if kvc%2 != 0 {
		Fatal(New(fmt.Sprintf(
			"errors: invalid argument count for WithFields, expected even number of fields, got %d",
			kvc,
		)))
	}

	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}

	for i := 0; i < kvc; i = i + 2 {
		key, ok := kvs[i].(string)
		if !ok {
			Fatal(New(fmt.Sprintf("errors: invalid type for key passed to WithFields at index %d", i)))
		}

		e.Fields[key] = kvs[i+1]
	}

	return e
}

// WithField appends a key/value pair to the error's field list.
func (e *Error) WithField(fieldKey string, fieldValue interface{}) *Error {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}

	e.Fields[fieldKey] = fieldValue

	return e
}

// format returns this error, and all previous errors, as a string. The result can be represented as
// a multi-line stack-trace by setting `asStack` to true.
func (e *Error) format(asStack bool) string {
	// Buffer is shared between recursive calls to avoid some unnecessary re-allocations.
	buf := bytes.Buffer{}

	e.formatAccumulator(&buf, asStack, false)

	return buf.String()
}

// formatAccumulator is a recursive error formatting function.
func (e *Error) formatAccumulator(buf *bytes.Buffer, asStack, isCause bool) {
	if asStack && !isCause {
		buf.WriteString("Error")
	}

	if e.caller != "" {
		pad(buf, ": ")
		buf.WriteString("[")
		buf.WriteString(e.caller)
		buf.WriteString("]")
	}

	if e.Message != "" {
		pad(buf, ": ")
		buf.WriteString(e.Message)
	}

	if e.Kind != "" {
		pad(buf, " ")
		buf.WriteString("(")
		buf.WriteString(string(e.Kind))
		buf.WriteString(")")
	}

	if asStack {
		buf.WriteString("\n")
		buf.WriteString("    ")
		buf.WriteString("File: \"")
		buf.WriteString(e.file)
		buf.WriteString("\", line ")
		buf.WriteString(strconv.Itoa(e.line))
		buf.WriteString("\n")

		if len(e.Fields) > 0 {
			buf.WriteString("    ")
			buf.WriteString("With fields:\n")

			fieldKeys := make([]string, 0, len(e.Fields))
			for k := range e.Fields {
				fieldKeys = append(fieldKeys, k)
			}

			sort.Strings(fieldKeys)

			for _, k := range fieldKeys {
				buf.WriteString("    ")
				buf.WriteString("- \"")
				buf.WriteString(k)
				buf.WriteString("\": ")
				buf.WriteString(fmt.Sprintf("%v", e.Fields[k]))
				buf.WriteString("\n")
			}
		}
	}

	if e.Cause != nil {
		if !asStack {
		} else {
			buf.WriteString("Caused by")
		}

		switch cause := e.Cause.(type) {
		case *Error:
			cause.formatAccumulator(buf, asStack, true)
		case error:
			pad(buf, ": ")
			buf.WriteString(cause.Error())
		}
	}
}

// pad takes a buffer and if it's not empty, writes the given padding string to it.
func pad(buf *bytes.Buffer, pad string) {
	if buf.Len() > 0 {
		buf.WriteString(pad)
	}
}

// New returns a new error. New accepts a variadic list of arguments, but at least one argument must
// be specified, otherwise New will panic. New will also panic if an unexpected type is given to it.
// Each field that can be set on an *Error is of a different type, meaning we can switch on the type
// of each argument, and still know which field to set on the error, leaving New as a very flexible
// function that is also not overly verbose to call.
//
// Example usage:
//
//    // Create the initial error, maybe this would be returned from some function.
//    err := errors.New(ErrKindTimeout, "client: HTTP request timed out")
//    // Wrap an existing error. It can be a regular error too. Also, set a field.
//    err = errors.New(err, "accom: fetch failed", errors.WithField("tti_code", ttiCode))
//
// As you can see, this usage is flexible, and includes the ability to construct pretty much any
// kind of error your application should need.
func New(args ...interface{}) *Error {
	err := newError(args...)

	updateCaller(err)

	return err
}

// newError creates a new *Error instance, returning it as an *Error, so that we can operate on it
// internally without having to cast back to *Error.
func newError(args ...interface{}) *Error {
	if len(args) == 0 {
		panic("errors: call to errors.New with no arguments")
	}

	err := &Error{}
	for _, arg := range args {
		switch v := arg.(type) {
		case Kind:
			err.Kind = v
		case string:
			err.Message = v
		case *Error:
			// Can't dereference a nil pointer, so bail early. This is a developer error.
			if v == nil {
				panic("errors: attempted to wrap nil *Error")
			}

			// Make a shallow copy of the value, so that we don't change the original error.
			cv := *v
			err.Cause = &cv
		case error:
			err.Cause = v
		case map[string]interface{}:
			err.Fields = v
		default:
			panic(fmt.Sprintf("errors: bad call to errors.New: unknown type %T, value %v", arg, arg))
		}
	}

	return err
}

// Wrap constructs an error the same way that New does, the only difference being that if the given
// cause is nil, this function will return nil. This makes it quite handy in return lines at the end
// of functions. Wrap conveys it's meaning a little more than New does when you are wrapping other
// errors.
func Wrap(cause error, args ...interface{}) *Error {
	if cause == nil {
		return nil
	}

	// Add the cause to the end of args so that it is definitely set as the cause.
	args = append(args, cause)
	err := newError(args...)

	// We have to set these again, as they'll be at the wrong depth now.
	updateCaller(err)

	return err
}

// updateCaller takes an error and sets the calling function information on it. Safe to use in error
// constructors, but no deeper.
func updateCaller(err *Error) {
	fpcs := make([]uintptr, 1)
	ptr := runtime.Callers(3, fpcs)
	if ptr == 0 {
		return
	}

	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun != nil {
		li := strings.LastIndex(fun.Name(), "/") + 1

		funcName := fun.Name()[li:]
		err.caller = funcName
		err.file, err.line = fun.FileLine(fpcs[0] - 1)
	}
}
