package flexbuffers

import (
	"math"
)

type value struct {
	d           int64
	typ         Type
	minBitWidth BitWidth
	hasExt      bool
}

func newValueBool(b bool) value {
	if b {
		return value{d: 1, typ: FBTBool, minBitWidth: 0}
	} else {
		return value{d: 0, typ: FBTBool, minBitWidth: 0}
	}
}

func (v value) StoredPackedType(bw BitWidth) uint8 {
	return PackedType(v.StoredWidth(bw), v.typ, v.hasExt)
}

func PaddingBytes(bufSize, scalarSize int) int {
	return ((^bufSize) + 1) & (scalarSize - 1)
}

func (v value) ElemWidth(bufSize, elemIndex int) BitWidth {
	if IsInline(v.typ) {
		return v.minBitWidth
	}
	// We have an absolute offset, but want to store a relative offset
	// elem_index elements beyond the current buffer end. Since whether
	// the relative offset fits in a certain byte_width depends on
	// the size of the elements before it (and their alignment), we have
	// to test for each size in turn.
	for byteWidth := 1; byteWidth <= 8; byteWidth *= 2 { // 1byte to 8 byte (uint8 to uint64)
		offsetLoc := bufSize + PaddingBytes(bufSize, byteWidth) + elemIndex*byteWidth
		offset := uint64(offsetLoc) - uint64(v.d)
		bitWidth := WidthU(offset)
		if 1<<bitWidth == byteWidth {
			return bitWidth
		}
	}
	panic("never here")
}

func (v value) StoredWidth(parentBitWidth BitWidth) BitWidth {
	if IsInline(v.typ) {
		if v.minBitWidth > parentBitWidth {
			return parentBitWidth
		} else {
			return v.minBitWidth
		}
	} else {
		return v.minBitWidth
	}
}

func (v value) AsFloat() float64 {
	if v.minBitWidth == BitWidth32 {
		return float64(math.Float32frombits(uint32(v.d)))
	} else {
		return math.Float64frombits(uint64(v.d))
	}
}

func (v value) AsInt() int64 {
	return v.d
}

func (v value) AsUInt() uint64 {
	return uint64(v.d)
}

func newValueUInt(u uint64, t Type, bw BitWidth, ext bool) value {
	return value{d: int64(u), typ: t, minBitWidth: bw, hasExt: ext}
}

func newValueInt(u int64, t Type, bw BitWidth) value {
	return value{d: u, typ: t, minBitWidth: bw}
}

func newValueFloat32(u float32) value {
	return value{d: int64(math.Float32bits(u)), typ: FBTFloat, minBitWidth: BitWidth32}
}

func newValueFloat64(d float64) value {
	// Refer: WidthF
	f := float32(d)
	if float64(f) == d {
		// float32
		return value{d: int64(math.Float32bits(f)), typ: FBTFloat, minBitWidth: BitWidth32}
	} else {
		// float64
		return value{d: int64(math.Float64bits(d)), typ: FBTFloat, minBitWidth: BitWidth64}
	}
}
