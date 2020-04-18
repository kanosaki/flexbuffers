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
				i, _ := w.BeginObject(nil)
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
				return w.EndObject(nil, i)
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
