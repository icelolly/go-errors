package errors

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	t.Run("should return the error in string form", func(t *testing.T) {
		err := New(Kind("testing"), "oops")
		assert.Equal(t, "[go-errors.TestError_Error.func1]: oops (testing)", err.Error())
	})
}

func TestError_Format(t *testing.T) {
	t.Run("should return the error in string form if formatted with %v", func(t *testing.T) {
		err := New(Kind("testing"), "oops")
		assert.Equal(t, "[go-errors.TestError_Format.func1]: oops (testing)", fmt.Sprintf("%v", err))
	})

	t.Run("should return the error in string form if formatted with %+v", func(t *testing.T) {
		err := Wrap(io.EOF, Kind("testing"), "oops").WithFields("foo", "bar")
		assert.Contains(t, fmt.Sprintf("%+v", err), "[go-errors.TestError_Format.func2]: oops (testing)")
		assert.Contains(t, fmt.Sprintf("%+v", err), "\n")
		assert.Contains(t, fmt.Sprintf("%+v", err), "File: ")
		assert.Contains(t, fmt.Sprintf("%+v", err), ", line ")
		assert.Contains(t, fmt.Sprintf("%+v", err), "foo")
		assert.Contains(t, fmt.Sprintf("%+v", err), "bar")
	})
}

func TestError_WithFields(t *testing.T) {
	t.Run("should attach the given fields to the error", func(t *testing.T) {
		err := New("oops").WithFields("foo", "bar", "baz", "qux")
		assert.Len(t, err.Fields, 2)

		_, fooOK := err.Fields["foo"]
		_, bazOK := err.Fields["foo"]

		assert.True(t, fooOK)
		assert.True(t, bazOK)
	})

	t.Run("should panic if an odd number of arguments is given", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = New("oops").WithFields("hello")
		})
	})

	t.Run("should panic if a non-string value is given as a key", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = New("oops").WithFields(1234, 1234)
		})
	})
}

func TestError_WithField(t *testing.T) {
	t.Run("should attach the given field to the error", func(t *testing.T) {
		err := New("oops").WithField("foo", "bar")
		assert.Len(t, err.Fields, 1)

		_, fooOK := err.Fields["foo"]

		assert.True(t, fooOK)
	})
}

func TestNew(t *testing.T) {
	t.Run("should not return nil", func(t *testing.T) {
		assert.NotNil(t, New("oops"))
	})

	t.Run("should panic if no arguments are passed", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = New()
		})
	})

	t.Run("should assign the given kind", func(t *testing.T) {
		kind := Kind("testing 1")
		err := New(kind)

		assert.Equal(t, kind, err.Kind)
	})

	t.Run("should assign the given message", func(t *testing.T) {
		message := "testing"
		err := New(message)

		assert.Equal(t, message, err.Message)
	})

	t.Run("should assign the given standard error cause", func(t *testing.T) {
		cause := errors.New("oops")
		err := New(cause)

		assert.Equal(t, cause, err.Cause)
	})

	t.Run("should assign the given go-errors error cause ", func(t *testing.T) {
		cause := New("oops")
		err := New(cause)

		assert.Equal(t, cause, err.Cause)
	})

	t.Run("should assign the given fields", func(t *testing.T) {
		err := New(map[string]interface{}{
			"foo": "bar",
			"baz": "qux",
		})

		assert.Len(t, err.Fields, 2)
	})

	t.Run("should panic if the given cause is a nil *Error", func(t *testing.T) {
		var cause *Error

		assert.Panics(t, func() {
			_ = New(cause)
		})
	})

	t.Run("should panic if the given argument is of an unexpected type", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = New(123)
		})
	})
}

func TestWrap(t *testing.T) {
	cause := New("oops")

	t.Run("should return nil if the given cause is nil", func(t *testing.T) {
		assert.Nil(t, Wrap(nil))
	})

	t.Run("should not return nil", func(t *testing.T) {
		assert.NotNil(t, Wrap(cause, "oops"))
	})

	t.Run("should assign the given kind", func(t *testing.T) {
		kind := Kind("testing 1")
		err := Wrap(cause, kind)

		assert.Equal(t, kind, err.Kind)
	})

	t.Run("should assign the given message", func(t *testing.T) {
		message := "testing"
		err := Wrap(cause, message)

		assert.Equal(t, message, err.Message)
	})

	t.Run("should assign the given standard error cause", func(t *testing.T) {
		cause := errors.New("oops")
		err := Wrap(cause)

		assert.Equal(t, cause, err.Cause)
	})

	t.Run("should assign the given go-errors error cause ", func(t *testing.T) {
		cause := New("oops")
		err := Wrap(cause)

		assert.Equal(t, cause, err.Cause)
	})

	t.Run("should assign the given fields", func(t *testing.T) {
		err := Wrap(cause, map[string]interface{}{
			"foo": "bar",
			"baz": "qux",
		})

		assert.Len(t, err.Fields, 2)
	})

	t.Run("should panic if the given cause is a nil *Error", func(t *testing.T) {
		var cause2 *Error

		assert.Panics(t, func() {
			_ = Wrap(cause, cause2)
		})
	})

	t.Run("should panic if the given argument is of an unexpected type", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = Wrap(cause, 123)
		})
	})
}

func BenchmarkNew(b *testing.B) {
	var err error

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = New(io.EOF, Kind("bench"), "benchmarking")
	}

	_ = err
}

func BenchmarkNewStd(b *testing.B) {
	var err error

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = errors.New("benchmarking")
	}

	_ = err
}
