package process

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"

	"flexbuffers"
)

type BSONReader struct {
	Data bson.Raw
}

func (b *BSONReader) FlushTo(ctx *Context, w DocumentWriter) error {
	return b.readDocument(ctx, w, b.Data)
}

func (b *BSONReader) readDocument(ctx *Context, output DocumentWriter, d bson.Raw) error {
	ptr, err := output.BeginObject(ctx)
	if err != nil {
		return err
	}
	elems, err := d.Elements()
	if err != nil {
		return err
	}
	for _, elem := range elems {
		key, err := elem.KeyErr()
		if err != nil {
			return err
		}
		if err := output.PushObjectKey(ctx, key); err != nil {
			return err
		}
		v, err := elem.ValueErr()
		if err != nil {
			return err
		}
		if err := b.readRawValue(ctx, output, v); err != nil {
			return err
		}
	}
	return output.EndObject(ctx, ptr)
}

func (b *BSONReader) readArray(ctx *Context, output DocumentWriter, d bson.Raw) error {
	ptr, err := output.BeginArray(ctx)
	if err != nil {
		return err
	}
	elems, err := d.Values()
	if err != nil {
		return err
	}
	for _, elem := range elems {
		if err := b.readRawValue(ctx, output, elem); err != nil {
			return err
		}
	}
	return output.EndArray(ctx, ptr)
}

func (b *BSONReader) readRawValue(ctx *Context, output DocumentWriter, rv bson.RawValue) error {
	switch rv.Type {
	case bsontype.Double:
		d, ok := rv.DoubleOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return output.PushFloat(ctx, d)
	case bsontype.String:
		v, ok := rv.StringValueOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return output.PushString(ctx, v)
	case bsontype.EmbeddedDocument:
		doc, ok := rv.DocumentOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.readDocument(ctx, output, doc)
	case bsontype.Array:
		v, ok := rv.ArrayOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.readArray(ctx, output, v)
	case bsontype.Binary:
		_, v, ok := rv.BinaryOK() // TODO: handle subtype
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return output.PushBlob(ctx, v)
	case bsontype.Boolean:
		v, ok := rv.BooleanOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return output.PushBool(ctx, v)
	case bsontype.Null:
		return output.PushNull(ctx)
	case bsontype.Int32:
		v, ok := rv.Int32OK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return output.PushInt(ctx, int64(v))
	case bsontype.Int64:
		v, ok := rv.Int64OK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return output.PushInt(ctx, v)
	case bsontype.DateTime:
		return fmt.Errorf("unsupported")
	case bsontype.Undefined:
		return fmt.Errorf("unsupported")
	case bsontype.ObjectID:
		return fmt.Errorf("unsupported")
	case bsontype.Regex:
		return fmt.Errorf("unsupported")
	case bsontype.DBPointer:
		return fmt.Errorf("unsupported")
	case bsontype.JavaScript:
		return fmt.Errorf("unsupported")
	case bsontype.Symbol:
		return fmt.Errorf("unsupported")
	case bsontype.CodeWithScope:
		return fmt.Errorf("unsupported")
	case bsontype.Timestamp:
		return fmt.Errorf("unsupported")
	case bsontype.Decimal128:
		return fmt.Errorf("unsupported")
	case bsontype.MinKey:
		return fmt.Errorf("unsupported")
	case bsontype.MaxKey:
		return fmt.Errorf("unsupported")
	default:
		return flexbuffers.ErrInvalidData
	}
}
