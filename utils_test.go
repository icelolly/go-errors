package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ErrKindTest Kind = "test"

func BenchmarkFields(b *testing.B) {
	err := New(ErrKindTest, "layer 1").WithField("foo", "bar")
	err = Wrap(err, "layer 2").WithField("bar", "baz")
	err = Wrap(err, "layer 3").WithField("baz", "qux")
	err = Wrap(err, "layer 4").WithField("qux", "quux")
	err = Wrap(err, "layer 5").WithField("bench", "mark")

	b.ReportAllocs()
	b.ResetTimer()

	var fields map[string]interface{}

	for i := 0; i < b.N; i++ {
		fields = Fields(err)
	}

	_ = fields
}

func BenchmarkStack(b *testing.B) {
	err := New(ErrKindTest, "layer 1")
	err = Wrap(err, "layer 2")
	err = Wrap(err, "layer 3")
	err = Wrap(err, "layer 4")
	err = Wrap(err, "layer 5")

	b.ReportAllocs()
	b.ResetTimer()

	var stack []StackFrame

	for i := 0; i < b.N; i++ {
		stack = Stack(err)
	}

	_ = stack
}

func TestFatal(t *testing.T) {
	t.Run("nil error should not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Fatal(nil)
		})
	})

	t.Run("error should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			Fatal(New("oops"))
		})
	})

	t.Run("error should contain message", func(t *testing.T) {
		message := "this is the error message"

		defer func() {
			if r := recover(); r != nil {
				str, ok := r.(string)
				require.True(t, ok)

				assert.True(t, strings.Contains(str, message))
			}
		}()

		Fatal(New(message))
	})

	t.Run("error should contain wrapped errors' messages", func(t *testing.T) {
		message1 := "this is the inner error message"
		message2 := "this is the middle error message"
		message3 := "this is the outer error message"

		defer func() {
			if r := recover(); r != nil {
				str, ok := r.(string)
				require.True(t, ok)

				assert.True(t, strings.Contains(str, message1))
				assert.True(t, strings.Contains(str, message2))
				assert.True(t, strings.Contains(str, message3))
			}
		}()

		Fatal(Wrap(Wrap(New(message1), message2), message3))
	})
}

func TestFields(t *testing.T) {
	t.Run("should return nil if the given error is nil", func(t *testing.T) {
		assert.Nil(t, Fields(nil))
	})

	t.Run("no fields on single error", func(t *testing.T) {
		err := New("oops")

		assert.Len(t, Fields(err), 0)
	})

	t.Run("no fields in stack of errors", func(t *testing.T) {
		err := New("oops")
		err = Wrap(err, "oops")
		err = Wrap(err, "oops")

		assert.Len(t, Fields(err), 0)
	})

	t.Run("one field on single error", func(t *testing.T) {
		err := New("oops").WithField("foo", "bar")

		flds := Fields(err)

		require.Len(t, flds, 1)
		assert.Equal(t, "bar", flds["foo"])
	})

	t.Run("one field at top of stack of errors", func(t *testing.T) {
		err := New("oops")
		err = Wrap(err, "oops").WithField("foo", "bar")

		flds := Fields(err)

		require.Len(t, flds, 1)
		assert.Equal(t, "bar", flds["foo"])
	})

	t.Run("multiple fields from multiple parts of stack of errors", func(t *testing.T) {
		err := New("oops").WithField("foo", "bar")
		err = Wrap(err, "oops").WithField("baz", "qux")

		flds := Fields(err)

		require.Len(t, flds, 2)
		assert.Equal(t, "bar", flds["foo"])
		assert.Equal(t, "qux", flds["baz"])
	})

	t.Run("same field multiple times in stack of errors", func(t *testing.T) {
		err := New("oops").WithField("foo", "qux")
		err = Wrap(err, "oops").WithField("foo", "bar")

		flds := Fields(err)

		require.Len(t, flds, 1)
		assert.Equal(t, "bar", flds["foo"])
	})

	t.Run("fields from stack that contains standard error", func(t *testing.T) {
		serr := errors.New("standard error")
		err := Wrap(serr, "oops").WithField("foo", "bar")

		flds := Fields(err)

		require.Len(t, flds, 1)
		assert.Equal(t, "bar", flds["foo"])
	})
}

func TestFieldsSlice(t *testing.T) {
	t.Run("should return nil if the given error is nil", func(t *testing.T) {
		assert.Nil(t, FieldsSlice(nil))
	})

	t.Run("no fields on single error", func(t *testing.T) {
		err := New("oops")

		assert.Len(t, FieldsSlice(err), 0)
	})

	t.Run("no fields in stack of errors", func(t *testing.T) {
		err := New("oops")
		err = Wrap(err, "oops")
		err = Wrap(err, "oops")

		assert.Len(t, FieldsSlice(err), 0)
	})

	t.Run("one field on single error", func(t *testing.T) {
		err := New("oops").WithField("foo", "bar")

		flds := FieldsSlice(err)

		require.Len(t, flds, 2)
		assert.Equal(t, "foo", flds[0])
		assert.Equal(t, "bar", flds[1])
	})

	t.Run("one field at top of stack of errors", func(t *testing.T) {
		err := New("oops")
		err = Wrap(err, "oops").WithField("foo", "bar")

		flds := FieldsSlice(err)

		require.Len(t, flds, 2)
		assert.Equal(t, "foo", flds[0])
		assert.Equal(t, "bar", flds[1])
	})

	t.Run("multiple fields from multiple parts of stack of errors", func(t *testing.T) {
		err := New("oops").WithField("foo", "bar")
		err = Wrap(err, "oops").WithField("baz", "qux")

		flds := FieldsSlice(err)

		require.Len(t, flds, 4)
		assert.Equal(t, "baz", flds[0])
		assert.Equal(t, "qux", flds[1])
		assert.Equal(t, "foo", flds[2])
		assert.Equal(t, "bar", flds[3])
	})

	t.Run("same field multiple times in stack of errors", func(t *testing.T) {
		err := New("oops").WithField("foo", "qux")
		err = Wrap(err, "oops").WithField("foo", "bar")

		flds := FieldsSlice(err)

		require.Len(t, flds, 2)
		assert.Equal(t, "foo", flds[0])
		assert.Equal(t, "bar", flds[1])
	})

	t.Run("fields from stack that contains standard error", func(t *testing.T) {
		serr := errors.New("standard error")
		err := Wrap(serr, "oops").WithField("foo", "bar")

		flds := FieldsSlice(err)

		require.Len(t, flds, 2)
		assert.Equal(t, "foo", flds[0])
		assert.Equal(t, "bar", flds[1])
	})
}

func TestIs(t *testing.T) {
	kind1 := Kind("testing 1")
	kind2 := Kind("testing 2")

	t.Run("should return false on nil error", func(t *testing.T) {
		assert.False(t, Is(nil))
	})

	t.Run("should return false if the error has no kind", func(t *testing.T) {
		err := New("oops")
		assert.False(t, Is(err, kind1))
	})

	t.Run("should return true if the error has the given kind", func(t *testing.T) {
		err := New(kind1)
		assert.True(t, Is(err, kind1))
	})

	t.Run("should return false if the error does not have the given kind", func(t *testing.T) {
		err := New(kind2)
		assert.False(t, Is(err, kind1))
	})

	t.Run("should return true if (only) a wrapped error has the given kind", func(t *testing.T) {
		err := Wrap(New(kind1))
		assert.True(t, Is(err, kind1))
	})

	t.Run("should return false if (only) a wrapped error has the given kind", func(t *testing.T) {
		err := Wrap(New(kind2))
		assert.False(t, Is(err, kind1))
	})

	t.Run("should test for the first of any given kinds", func(t *testing.T) {
		err := Wrap(New(kind1))
		assert.True(t, Is(err, kind2, kind1))
	})
}

func TestMessage(t *testing.T) {
	fallback := "An internal error has occurred. Please contact technical support."

	t.Run("should return an empty message on nil error", func(t *testing.T) {
		assert.Equal(t, Message(nil), "")
	})

	t.Run("should return a default string if no message is set", func(t *testing.T) {
		err := New(Kind("test"))
		assert.Equal(t, Message(err), fallback)
	})

	t.Run("should return the message set on the error", func(t *testing.T) {
		err := New("oops")
		assert.Equal(t, Message(err), "oops")
	})

	t.Run("should return a deeply nested message if no direct message is set", func(t *testing.T) {
		err := Wrap(New("oops"))
		assert.Equal(t, Message(err), "oops")
	})
}

func TestStack(t *testing.T) {
	kind1 := Kind("testing 1")
	kind2 := Kind("testing 2")

	t.Run("should return an empty stick on nil error", func(t *testing.T) {
		assert.Len(t, Stack(nil), 0)
	})

	t.Run("should return a stack with 1 item for a new error", func(t *testing.T) {
		assert.Len(t, Stack(New("oops")), 1)
	})

	t.Run("should return a stack with 2 item for a wrapped error with a single cause", func(t *testing.T) {
		assert.Len(t, Stack(Wrap(New("oops"))), 2)
	})

	t.Run("should work whilst wrapping standard errors", func(t *testing.T) {
		assert.Len(t, Stack(Wrap(errors.New("oops"))), 2)
	})

	t.Run("should contain the kind for each error", func(t *testing.T) {
		stack := Stack(Wrap(New(kind1), kind2))

		assert.Equal(t, string(kind2), stack[0].Kind)
		assert.Equal(t, string(kind1), stack[1].Kind)
	})

	t.Run("should contain the message for each error", func(t *testing.T) {
		stack := Stack(Wrap(New("testing 1"), "testing 2"))

		assert.Equal(t, "testing 2", stack[0].Message)
		assert.Equal(t, "testing 1", stack[1].Message)
	})

	t.Run("should contain the fields for each error", func(t *testing.T) {
		err1 := New("testing 1").WithFields("foo", "bar", "baz", "qux")
		err2 := Wrap(err1).WithFields("lorem", "ipsum")

		stack := Stack(err2)

		assert.Len(t, stack[0].Fields, 1)
		assert.Len(t, stack[1].Fields, 2)
	})
}
