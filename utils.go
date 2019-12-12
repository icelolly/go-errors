package errors

import (
	"fmt"
	"sort"
)

// Fatal will panic if given a non-nil error. If the given error is an *Error, the output format of
// the panic will be slightly different, so as to include as much relevant information as possible
// in an easy format for operators to digest. If a regular error is given, it will simply be passed
// to panic as normal.
func Fatal(err error) {
	if err == nil {
		return
	}

	wrapped := newError(err)

	updateCaller(wrapped)

	if v, ok := err.(*Error); ok {
		isFileMatch := v.file == wrapped.file
		isLineMatch := v.line == wrapped.line
		isCallerMatch := v.caller == wrapped.caller

		if isFileMatch && isLineMatch && isCallerMatch {
			// If the error we've just created to try to wrap the error that was passed has the same
			// file, line, and caller, then we've just duplicated a stack frame, so let's throw that
			// away instead of showing that in the panic output.
			wrapped = v
		}
	}

	panic(fmt.Sprintf("fatal error: %s\n\n%+v", Message(wrapped), wrapped))
}

// Fields returns all fields from all errors in a stack of errors, recursively checking for fields
// and merging them into one map, then returning them.
func Fields(err error) map[string]interface{} {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if !ok {
		return nil
	}

	var fields map[string]interface{}

	if e.Cause != nil {
		causeFields := Fields(e.Cause)
		if causeFields != nil {
			// Try our best to avoid growth allocations...
			fields = make(map[string]interface{}, len(e.Fields)+len(causeFields))

			for k, v := range causeFields {
				fields[k] = v
			}
		}
	}

	if e.Fields != nil {
		if fields == nil {
			// We haven't got any fields from other parts of the stack, so just use the fields on
			// this error instead of allocating more memory.
			fields = e.Fields
		} else {
			// Otherwise, we did find errors in another part of the stack, and because of that,
			// fields shouldn't be nil, and should be the right size.
			for k, v := range e.Fields {
				fields[k] = v
			}
		}
	}

	return fields
}

// FieldsSlice returns all fields from all errors in a stack of errors, recursively checking for
// fields and merging them into one slice, then returning them. This function uses Fields
// internally, so the behaviour is very similar. The returned slice is ordered by key, so calling
// this function produces consistent results.
func FieldsSlice(err error) []interface{} {
	if err == nil {
		return nil
	}

	fieldMap := Fields(err)

	var fieldKeys []string
	for k := range fieldMap {
		fieldKeys = append(fieldKeys, k)
	}

	sort.Strings(fieldKeys)

	var fields []interface{}
	for _, k := range fieldKeys {
		fields = append(fields, k)
		fields = append(fields, fieldMap[k])
	}

	return fields
}

// Is reports whether the err is an *Error of the given kind/value. If the given kind is of type Kind/string, it will be
// checked against the error's Kind. If the given kind is of any other type, it will be checked against the error's
// cause. This is done recursively until a matching error is found. Calling Is with multiple kinds reports whether the
// error is one of the given kind/values, not all of.
func Is(err error, kind ...interface{}) bool {
	if err == nil {
		return false
	}

	e, ok := err.(*Error)
	if !ok {
		return false
	}

	for _, k := range kind {
		switch val := k.(type) {
		case Kind, string:
			if e.Kind == val {
				return true
			}
		default:
			if e.Cause == val {
				return true
			}
		}
	}

	if e.Cause != nil {
		return Is(e.Cause, kind...)
	}

	return false
}

// Message returns what is supposed to be a human-readable error message. It is designed to not leak
// internal implementation details (unlike calling *Error.Error()). If the given error is not an
// *Error, then a generic message will be returned. If the given error is nil, then an empty string
// will be returned.
func Message(err error) string {
	if err == nil {
		return ""
	}

	e, ok := err.(*Error)
	if ok && e.Message != "" {
		return e.Message
	} else if ok && e.Cause != nil {
		return Message(e.Cause)
	}

	return "An internal error has occurred. Please contact technical support."
}

// StackFrame represents a single error in a stack of errors. All fields could be empty, because we
// may even be dealing with a regular error.
type StackFrame struct {
	Kind    string                 `json:"kind,omitempty"`
	Message string                 `json:"message,omitempty"`
	Caller  string                 `json:"caller,omitempty"`
	File    string                 `json:"file,omitempty"`
	Line    int                    `json:"line,omitempty"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// Stack produces a slice of StackFrame structs that can easily be encoded to JSON. The main
// intended use of this function is for logging, so that you can attach a stack trace to a log entry
// to help track down the cause of an error.
//
// This function looks a little more complex than some of the other recursive alternatives, but
// because of the nature of slices, this implementation is considerably faster than using a
// recursive solution (i.e. this only has 1 allocation, whereas a recursive solution may have 1 or 2
// allocations per stack frame).
func Stack(err error) []StackFrame {
	if err == nil {
		return []StackFrame{}
	}

	original := err

	// Calculate the size of the slice we should make by finding the size of the stack.
	var size int
	for err != nil {
		size++

		e, ok := err.(*Error)
		if !ok {
			break
		}

		err = e.Cause
	}

	// Reset err to the original err, not the end of the stack.
	err = original

	// Produce a slice of StackFrame that should not need to grow, avoiding unnecessary allocations.
	stack := make([]StackFrame, 0, size)

	for err != nil {
		// If we don't see an *Error, we must be at the end, and should just return a stack frame that
		// just contains the error's message.
		e, ok := err.(*Error)
		if !ok {
			stack = append(stack, StackFrame{
				Message: err.Error(),
			})
			break
		}

		// Produce a stack frame for this *Error.
		stack = append(stack, StackFrame{
			Kind:    string(e.Kind),
			Message: e.Message,
			Fields:  e.Fields,
			Caller:  e.caller,
			File:    e.file,
			Line:    e.line,
		})

		// Set err to the next error in the stack. If it's nil, the loop condition will break.
		err = e.Cause
	}

	return stack
}
