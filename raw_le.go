//+build amd64

package flexbuffers

import (
	"math"
	"unsafe"
)

type Raw []byte

func (b Raw) ReadUInt64(offset int, byteWidth uint8) uint64 {
	if byteWidth < 4 {
		if byteWidth < 2 {
			return uint64(*(*uint8)(unsafe.Pointer(&b[offset])))
		} else {
			return uint64(*(*uint16)(unsafe.Pointer(&b[offset])))
		}
	} else {
		if byteWidth < 8 {
			return uint64(*(*uint32)(unsafe.Pointer(&b[offset])))
		} else {
			return uint64(*(*uint64)(unsafe.Pointer(&b[offset])))
		}
	}
}

func (b Raw) ReadInt64(offset int, byteWidth uint8) int64 {
	if byteWidth < 4 {
		if byteWidth < 2 {
			return int64(*(*int8)(unsafe.Pointer(&b[offset])))
		} else {
			return int64(*(*int16)(unsafe.Pointer(&b[offset])))
		}
	} else {
		if byteWidth < 8 {
			return int64(*(*int32)(unsafe.Pointer(&b[offset])))
		} else {
			return int64(*(*int64)(unsafe.Pointer(&b[offset])))
		}
	}
}

func (b Raw) ReadDouble(offset int, byteWidth uint8) float64 {
	if byteWidth < 4 {
		if byteWidth < 2 {
			panic("float8 is not supported")
		} else {
			panic("float16 is not supported")
		}
	} else {
		if byteWidth < 8 {
			return float64(*(*float32)(unsafe.Pointer(&b[offset])))
		} else {
			return *(*float64)(unsafe.Pointer(&b[offset]))
		}
	}
}

func (b Raw) WriteInt64(offset int, byteWidth uint8, value int64) bool {
	valueWidth := WidthI(value)
	fits := (1 << valueWidth) <= byteWidth
	if fits {
		if valueWidth == BitWidth8 {
			*(*int8)(unsafe.Pointer(&b[offset])) = int8(value)
		} else if valueWidth == BitWidth16 {
			*(*int16)(unsafe.Pointer(&b[offset])) = int16(value)
		} else if valueWidth == BitWidth32 {
			*(*int32)(unsafe.Pointer(&b[offset])) = int32(value)
		} else {
			*(*int64)(unsafe.Pointer(&b[offset])) = value
		}
	}
	return fits
}

func (b Raw) WriteUInt64(offset int, byteWidth uint8, value uint64) bool {
	valueWidth := WidthU(value)
	fits := (1 << valueWidth) <= byteWidth
	if fits {
		if valueWidth == BitWidth8 {
			*(*uint8)(unsafe.Pointer(&b[offset])) = uint8(value)
		} else if valueWidth == BitWidth16 {
			*(*uint16)(unsafe.Pointer(&b[offset])) = uint16(value)
		} else if valueWidth == BitWidth32 {
			*(*uint32)(unsafe.Pointer(&b[offset])) = uint32(value)
		} else {
			*(*uint64)(unsafe.Pointer(&b[offset])) = value
		}
	}
	return fits
}

func (b Raw) WriteFloat(offset int, byteWidth uint8, value float64) bool {
	valueWidth := WidthF(value)
	fits := (1 << valueWidth) <= byteWidth
	if !fits {
		return false
	}
	if byteWidth == 4 {
		*(*uint32)(unsafe.Pointer(&b[offset])) = math.Float32bits(float32(value))
	} else if byteWidth == 8 {
		*(*uint64)(unsafe.Pointer(&b[offset])) = math.Float64bits(value)
	}
	return true
}

func (b Raw) Indirect(offset int, byteWidth uint8) int {
	if byteWidth < 4 {
		if byteWidth < 2 {
			return offset - int(*(*uint8)(unsafe.Pointer(&b[offset])))
		} else {
			return offset - int(*(*uint16)(unsafe.Pointer(&b[offset])))
		}
	} else {
		if byteWidth < 8 {
			return offset - int(*(*uint32)(unsafe.Pointer(&b[offset])))
		} else {
			return offset - int(*(*uint64)(unsafe.Pointer(&b[offset])))
		}
	}
}

func (b Raw) Root() Reference {
	_ = b[len(b)-3] // check boundary
	byteWidth := b[len(b)-1]
	packedType := b[len(b)-2]
	rootOffset := len(b) - 2 - int(byteWidth)
	return NewReferenceFromPackedType(b, rootOffset, byteWidth, packedType)
}

func (b Raw) InitTraverser(tv *Traverser) {
	byteWidth := b[len(b)-1]
	packedType := b[len(b)-2]
	rootOffset := len(b) - 2 - int(byteWidth)
	*tv = Traverser{
		buf:         b,
		offset:      rootOffset,
		typ:         Type(packedType >> 2),
		parentWidth: int(byteWidth),
		byteWidth:   1 << (packedType & 3),
	}
}

func (b Raw) Lookup(path ...string) Reference {
	var tv Traverser
	b.InitTraverser(&tv)
	tv.Seek(path)
	return tv.Current()
}
