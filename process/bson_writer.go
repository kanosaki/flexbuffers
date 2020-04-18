package process

import (
	"fmt"
	"math"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type BSONWriter struct {
	dst []byte
	key string

	elemIndex []int
}

func (w *BSONWriter) pushHeader(t bsontype.Type) error {
	if w.key == "" && len(w.elemIndex) == 0 {
		// non-object && non-array --> no header
		return nil
	}
	key := w.key
	if key == "" {
		key = strconv.Itoa(w.elemIndex[len(w.elemIndex)-1]) // TODO: cache?
		w.elemIndex[len(w.elemIndex)-1]++
	}
	w.dst = bsoncore.AppendHeader(w.dst, t, key)
	w.key = ""
	return nil
}

func (w *BSONWriter) PushString(ctx *Context, s string) error {
	if err := w.pushHeader(bsontype.String); err != nil {
		return err
	}
	w.dst = bsoncore.AppendString(w.dst, s)
	return nil
}

func (w *BSONWriter) PushBlob(ctx *Context, b []byte) error {
	if err := w.pushHeader(bsontype.Binary); err != nil {
		return err
	}
	w.dst = bsoncore.AppendBinary(w.dst, 0, b)
	return nil
}

func (w *BSONWriter) PushInt(ctx *Context, i int64) error {
	if i < math.MinInt64 || math.MaxInt64 < i {
		return fmt.Errorf("cannot write %d to BSON document: out of bound", i)
	}
	if err := w.pushHeader(bsontype.Int64); err != nil {
		return err
	}
	w.dst = bsoncore.AppendInt64(w.dst, i)
	return nil
}

func (w *BSONWriter) PushUint(ctx *Context, u uint64) error {
	if math.MaxInt64 < u {
		return fmt.Errorf("cannot write %d to BSON document: out of bound", u)
	}
	if err := w.pushHeader(bsontype.Int64); err != nil {
		return err
	}
	w.dst = bsoncore.AppendInt64(w.dst, int64(u))
	return nil
}

func (w *BSONWriter) PushFloat(ctx *Context, f float64) error {
	if err := w.pushHeader(bsontype.Double); err != nil {
		return err
	}
	w.dst = bsoncore.AppendDouble(w.dst, f)
	return nil
}

func (w *BSONWriter) PushBool(ctx *Context, b bool) error {
	if err := w.pushHeader(bsontype.Boolean); err != nil {
		return err
	}
	w.dst = bsoncore.AppendBoolean(w.dst, b)
	return nil
}

func (w *BSONWriter) PushNull(*Context) error {
	if err := w.pushHeader(bsontype.Null); err != nil {
		return err
	}
	return nil
}

func (w *BSONWriter) BeginArray(*Context) (int, error) {
	if err := w.pushHeader(bsontype.Array); err != nil {
		return 0, err
	}
	w.elemIndex = append(w.elemIndex, 0)
	var ptr int32
	ptr, w.dst = bsoncore.AppendDocumentStart(w.dst)
	return int(ptr), nil
}

func (w *BSONWriter) EndArray(ctx *Context, id int) error {
	w.elemIndex = w.elemIndex[:len(w.elemIndex)-1]
	var err error
	w.dst, err = bsoncore.AppendDocumentEnd(w.dst, int32(id))
	return err
}

func (w *BSONWriter) BeginObject(*Context) (int, error) {
	if err := w.pushHeader(bsontype.EmbeddedDocument); err != nil {
		return 0, err
	}
	w.elemIndex = append(w.elemIndex, 0)
	var ptr int32
	ptr, w.dst = bsoncore.AppendDocumentStart(w.dst)
	return int(ptr), nil
}

func (w *BSONWriter) EndObject(ctx *Context, id int) error {
	w.key = ""
	w.elemIndex = w.elemIndex[:len(w.elemIndex)-1]
	var err error
	w.dst, err = bsoncore.AppendDocumentEnd(w.dst, int32(id))
	return err
}

func (w *BSONWriter) PushObjectKey(ctx *Context, k string) error {
	w.key = k
	return nil
}
