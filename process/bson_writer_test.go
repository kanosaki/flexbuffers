package process

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBsonWriter(t *testing.T) {
	cases := []struct {
		fn       func(w *BSONWriter) error
		expected bson.D
	}{
		{
			fn: func(w *BSONWriter) error {
				i, _ := w.BeginObject()
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
				return w.EndObject(i)
			},
			expected: bson.D{
				{Key: "a", Value: "foo"},
				{Key: "b", Value: int64(-1)},
				{Key: "c", Value: int64(1)},
				{Key: "d", Value: 1.234},
				{Key: "e", Value: true},
				{Key: "f", Value: false},
				{Key: "g", Value: nil},
			},
		},
		{
			fn: func(w *BSONWriter) error {
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
			expected: bson.D{
				{
					Key: "a",
					Value: bson.D{
						{Key: "x", Value: "y"},
					},
				},
				{
					Key: "b",
					Value: bson.A{
						int64(1),
						int64(2),
						bson.D{
							{Key: "x", Value: "y"},
						},
					},
				},
			},
		},
	}
	for _, cas := range cases {
		var dst []byte
		w := &BSONWriter{dst: dst}
		if err := cas.fn(w); err != nil {
			t.Fatal(err)
		}
		actual := bson.D{}
		if err := bson.Unmarshal(w.dst, &actual); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(cas.expected, actual); diff != "" {
			t.Error(diff)
		}
	}
}
