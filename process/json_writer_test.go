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
				_, _ = w.BeginObject(nil)
				return w.EndObject(nil, 0)
			},
			expected: "{}",
		},
		{
			fn: func(w *JsonWriter) error {
				_, _ = w.BeginObject(nil)
				_ = w.PushObjectKey(nil, "a")
				_ = w.PushString(nil, "foo")
				_ = w.PushObjectKey(nil, "b")
				_ = w.PushInt(nil, -1)
				_ = w.PushObjectKey(nil, "c")
				_ = w.PushUint(nil, 1)
				_ = w.PushObjectKey(nil, "d")
				_ = w.PushFloat(nil, 1.234)
				_ = w.PushObjectKey(nil, "e")
				_ = w.PushBool(nil, true)
				_ = w.PushObjectKey(nil, "f")
				_ = w.PushBool(nil, false)
				_ = w.PushObjectKey(nil, "g")
				_ = w.PushNull(nil)
				return w.EndObject(nil, 0)
			},
			expected: `{"a":"foo","b":-1,"c":1,"d":1.234000,"e":true,"f":false,"g":null}`,
		},
		{
			fn: func(w *JsonWriter) error {
				m, _ := w.BeginObject(nil)
				_ = w.PushObjectKey(nil, "a")
				n, _ := w.BeginObject(nil)
				_ = w.PushObjectKey(nil, "x")
				_ = w.PushString(nil, "y")
				_ = w.EndObject(nil, n)
				_ = w.PushObjectKey(nil, "b")
				o, _ := w.BeginArray(nil)
				_ = w.PushUint(nil, 1)
				_ = w.PushInt(nil, 2)
				p, _ := w.BeginObject(nil)
				_ = w.PushObjectKey(nil, "x")
				_ = w.PushString(nil, "y")
				_ = w.EndObject(nil, p)
				_ = w.EndArray(nil, o)
				return w.EndObject(nil, m)
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
