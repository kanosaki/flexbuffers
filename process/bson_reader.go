package process

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"

	"flexbuffers"
)

type BSONReader struct {
	Output DocumentWriter
}

func (b *BSONReader) readDocument(d bson.Raw) error {
	ptr, err := b.Output.BeginObject()
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
		if err := b.Output.PushObjectKey(key); err != nil {
			return err
		}
		v, err := elem.ValueErr()
		if err != nil {
			return err
		}
		if err := b.readRawValue(v); err != nil {
			return err
		}
	}
	return b.Output.EndObject(ptr)
}

func (b *BSONReader) readArray(d bson.Raw) error {
	ptr, err := b.Output.BeginArray()
	if err != nil {
		return err
	}
	elems, err := d.Values()
	if err != nil {
		return err
	}
	for _, elem := range elems {
		if err := b.readRawValue(elem); err != nil {
			return err
		}
	}
	return b.Output.EndArray(ptr)
}

func (b *BSONReader) readRawValue(rv bson.RawValue) error {
	switch rv.Type {
	case bsontype.Double:
		d, ok := rv.DoubleOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.Output.PushFloat(d)
	case bsontype.String:
		v, ok := rv.StringValueOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.Output.PushString(v)
	case bsontype.EmbeddedDocument:
		doc, ok := rv.DocumentOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.readDocument(doc)
	case bsontype.Array:
		v, ok := rv.ArrayOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.readArray(v)
	case bsontype.Binary:
		_, v, ok := rv.BinaryOK() // TODO: handle subtype
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.Output.PushBlob(v)
	case bsontype.Boolean:
		v, ok := rv.BooleanOK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.Output.PushBool(v)
	case bsontype.Null:
		return b.Output.PushNull()
	case bsontype.Int32:
		v, ok := rv.Int32OK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.Output.PushInt(int64(v))
	case bsontype.Int64:
		v, ok := rv.Int64OK()
		if !ok {
			return flexbuffers.ErrInvalidData
		}
		return b.Output.PushInt(v)
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
