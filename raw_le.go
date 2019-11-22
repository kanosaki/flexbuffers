//+build unsafe

package flexbuffers

import (
	"errors"
	"fmt"
	"math"
	"unsafe"
)

var (
	ErrUpdateDoesntFit = errors.New("update data doesn't fit")
)

type Raw []byte

func (b Raw) ReadUInt64(offset int, byteWidth uint8) (uint64, error) {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return 0, ErrOutOfRange
	}
	if byteWidth < 4 {
		if byteWidth < 2 {
			return uint64(*(*uint8)(unsafe.Pointer(&b[offset]))), nil
		} else {
			return uint64(*(*uint16)(unsafe.Pointer(&b[offset]))), nil
		}
	} else {
		if byteWidth < 8 {
			return uint64(*(*uint32)(unsafe.Pointer(&b[offset]))), nil
		} else {
			return uint64(*(*uint64)(unsafe.Pointer(&b[offset]))), nil
		}
	}
}

func (b Raw) ReadInt64(offset int, byteWidth uint8) (int64, error) {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return 0, ErrOutOfRange
	}
	if byteWidth < 4 {
		if byteWidth < 2 {
			return int64(*(*int8)(unsafe.Pointer(&b[offset]))), nil
		} else {
			return int64(*(*int16)(unsafe.Pointer(&b[offset]))), nil
		}
	} else {
		if byteWidth < 8 {
			return int64(*(*int32)(unsafe.Pointer(&b[offset]))), nil
		} else {
			return int64(*(*int64)(unsafe.Pointer(&b[offset]))), nil
		}
	}
}

func (b Raw) ReadDouble(offset int, byteWidth uint8) (float64, error) {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return 0.0, ErrOutOfRange
	}
	if byteWidth < 4 {
		if byteWidth < 2 {
			return 0.0, fmt.Errorf("float8 is not supported")
		} else {
			return 0.0, fmt.Errorf("float16 is not supported")
		}
	} else {
		if byteWidth < 8 {
			return float64(*(*float32)(unsafe.Pointer(&b[offset]))), nil
		} else {
			return *(*float64)(unsafe.Pointer(&b[offset])), nil
		}
	}
}

func (b Raw) WriteInt64(offset int, byteWidth uint8, value int64) error {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return ErrOutOfRange
	}
	valueWidth := WidthI(value)
	fits := (1 << valueWidth) <= byteWidth
	if !fits {
		return ErrUpdateDoesntFit
	}
	if valueWidth == BitWidth8 {
		*(*int8)(unsafe.Pointer(&b[offset])) = int8(value)
	} else if valueWidth == BitWidth16 {
		*(*int16)(unsafe.Pointer(&b[offset])) = int16(value)
	} else if valueWidth == BitWidth32 {
		*(*int32)(unsafe.Pointer(&b[offset])) = int32(value)
	} else {
		*(*int64)(unsafe.Pointer(&b[offset])) = value
	}
	return nil
}

func (b Raw) WriteUInt64(offset int, byteWidth uint8, value uint64) error {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return ErrOutOfRange
	}
	valueWidth := WidthU(value)
	fits := (1 << valueWidth) <= byteWidth
	if !fits {
		return ErrUpdateDoesntFit
	}
	if valueWidth == BitWidth8 {
		*(*uint8)(unsafe.Pointer(&b[offset])) = uint8(value)
	} else if valueWidth == BitWidth16 {
		*(*uint16)(unsafe.Pointer(&b[offset])) = uint16(value)
	} else if valueWidth == BitWidth32 {
		*(*uint32)(unsafe.Pointer(&b[offset])) = uint32(value)
	} else {
		*(*uint64)(unsafe.Pointer(&b[offset])) = value
	}
	return nil
}

func (b Raw) WriteFloat(offset int, byteWidth uint8, value float64) error {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return ErrOutOfRange
	}
	valueWidth := WidthF(value)
	fits := (1 << valueWidth) <= byteWidth
	if !fits {
		return ErrUpdateDoesntFit
	}
	if byteWidth == 4 {
		*(*uint32)(unsafe.Pointer(&b[offset])) = math.Float32bits(float32(value))
	} else if byteWidth == 8 {
		*(*uint64)(unsafe.Pointer(&b[offset])) = math.Float64bits(value)
	}
	return nil
}

func (b Raw) Indirect(offset int, byteWidth uint8) (int, error) {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return 0, ErrOutOfRange
	}
	var ind int
	if byteWidth < 4 {
		if byteWidth < 2 {
			ind = offset - int(*(*uint8)(unsafe.Pointer(&b[offset])))
		} else {
			ind = offset - int(*(*uint16)(unsafe.Pointer(&b[offset])))
		}
	} else {
		if byteWidth < 8 {
			ind = offset - int(*(*uint32)(unsafe.Pointer(&b[offset])))
		} else {
			ind = offset - int(*(*uint64)(unsafe.Pointer(&b[offset])))
		}
	}
	if ind < 0 || len(b) <= ind {
		return 0, ErrOutOfRange
	}
	return ind, nil
}

func (b Raw) Validate() error {
	root, err := b.Root()
	if err != nil {
		return err
	}
	return root.Validate()
}

func (b Raw) RootOrNull() Reference {
	v, err := b.Root()
	if err != nil {
		return NullReference
	}
	return v
}

func (b Raw) Root() (Reference, error) {
	if len(b) <= 2 {
		return Reference{}, ErrInvalidData
	}
	byteWidth := b[len(b)-1]
	packedType := b[len(b)-2]
	rootOffset := len(b) - 2 - int(byteWidth)
	return NewReferenceFromPackedType(&b, rootOffset, byteWidth, packedType)
}

func (b Raw) InitTraverser(tv *Traverser) {
	byteWidth := b[len(b)-1]
	packedType := b[len(b)-2]
	rootOffset := len(b) - 2 - int(byteWidth)
	*tv = Traverser{
		buf:         &b,
		offset:      rootOffset,
		typ:         Type(packedType >> 2),
		parentWidth: int(byteWidth),
		byteWidth:   1 << (packedType & 3),
	}
}

func (b Raw) LookupOrNull(path ...string) Reference {
	r, err := b.Lookup(path...)
	if err != nil {
		return NullReference
	}
	return r
}

func (b Raw) Lookup(path ...string) (Reference, error) {
	var tv Traverser
	b.InitTraverser(&tv)
	if err := tv.Seek(path); err != nil {
		return Reference{}, err
	}
	return tv.Current()
}
