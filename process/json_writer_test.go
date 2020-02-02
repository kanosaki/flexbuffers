package process

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestJsonWriter(t *testing.T) {
	cases := []struct {
		fn       func(w *JsonWriter) error
		expected string
	}{
		{
			fn: func(w *JsonWriter) error {
				_, _ = w.BeginObject()
				return w.EndObject(0)
			},
			expected: "{}",
		},
		{
			fn: func(w *JsonWriter) error {
				_, _ = w.BeginObject()
				_ = w.PushObjectKey("a")
				_ = w.PushString("foo")
				_ = w.PushObjectKey("b")
				_ = w.PushInt(-1)
				_ = w.PushObjectKey("c")
				_ = w.PushUint(1)
				_ = w.PushObjectKey("d")
				_ = w.PushFloat(1.234)
				_ = w.PushObjectKey("e")
				_ = w.PushBool(true)
				_ = w.PushObjectKey("f")
				_ = w.PushBool(false)
				_ = w.PushObjectKey("g")
				_ = w.PushNull()
				return w.EndObject(0)
			},
			expected: `{"a":"foo","b":-1,"c":1,"d":1.234000,"e":true,"f":false,"g":null}`,
		},
		{
			fn: func(w *JsonWriter) error {
				m, _ := w.BeginObject()
				_ = w.PushObjectKey("a")
				n, _ := w.BeginObject()
				_ = w.PushObjectKey("x")
				_ = w.PushString("y")
				_ = w.EndObject(n)
				_ = w.PushObjectKey("b")
				o, _ := w.BeginArray()
				_ = w.PushUint(1)
				_ = w.PushInt(2)
				p, _ := w.BeginObject()
				_ = w.PushObjectKey("x")
				_ = w.PushString("y")
				_ = w.EndObject(p)
				_ = w.EndArray(o)
				return w.EndObject(m)
			},
			expected: `{"a":{"x":"y"},"b":[1,2,{"x":"y"}]}`,
		},
	}
	for _, cas := range cases {
		var buf bytes.Buffer
		w := &JsonWriter{Output: &buf}
		if err := cas.fn(w); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(cas.expected, buf.String()); diff != "" {
			t.Error(diff)
		}
	}
}
