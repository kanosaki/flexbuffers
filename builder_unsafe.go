//+build unsafe

package flexbuffers

import (
	"fmt"
	"unsafe"
)

func (b *Builder) IndirectInt(i int64) {
	var tmp [8]byte
	bitWidth := WidthI(i)
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	*((*int64)(unsafe.Pointer(&tmp[0]))) = i
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectInt, bitWidth))
}

func (b *Builder) IndirectUInt(i uint64) {
	var tmp [8]byte
	bitWidth := WidthU(i)
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	*((*uint64)(unsafe.Pointer(&tmp[0]))) = i
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectUInt, bitWidth))
}

func (b *Builder) IndirectFloat32(f float32) {
	var tmp [4]byte
	bitWidth := BitWidth32
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	*((*float32)(unsafe.Pointer(&tmp[0]))) = f
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectFloat, bitWidth))
}

func (b *Builder) IndirectFloat64(f float64) {
	var tmp [8]byte
	bitWidth := WidthF(f)
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	*((*float64)(unsafe.Pointer(&tmp[0]))) = f
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectFloat, bitWidth))
}

func (b *Builder) allocateUnsafe(bw int) unsafe.Pointer {
	l := len(b.buf)
	if bw <= cap(b.buf)-l {
		b.buf = b.buf[:l+bw]
	} else {
		b.buf = append(b.buf, make([]byte, bw)...)
	}
	return unsafe.Pointer(&b.buf[l])
}

func (b *Builder) WriteUInt(i uint64, bw int) {
	ptr := b.allocateUnsafe(bw)
	if bw == 1 {
		*(*uint8)(ptr) = uint8(i)
	} else if bw == 2 {
		*(*uint16)(ptr) = uint16(i)
	} else if bw == 4 {
		*(*uint32)(ptr) = uint32(i)
	} else if bw == 8 {
		*(*uint64)(ptr) = uint64(i)
	}
}

func (b *Builder) WriteInt(i int64, bw int) {
	ptr := b.allocateUnsafe(bw)
	if bw == 1 {
		*(*int8)(ptr) = int8(i)
	} else if bw == 2 {
		*(*int16)(ptr) = int16(i)
	} else if bw == 4 {
		*(*int32)(ptr) = int32(i)
	} else if bw == 8 {
		*(*int64)(ptr) = int64(i)
	}
}
func (b *Builder) WriteDouble(f float64, byteWidth int) error {
	ptr := b.allocateUnsafe(byteWidth)
	if byteWidth == 4 {
		*(*float32)(ptr) = float32(f)
	} else if byteWidth == 8 {
		*(*float64)(ptr) = f
	} else {
		return fmt.Errorf("invalid float width: %d", byteWidth)
	}
	return nil
}

func (b *Builder) scalarVector(elems []interface{}, fixed bool) (int, error) {
	vectorType, err := GetScalarType(elems[0])
	if err != nil {
		return 0, err
	}
	l := len(elems)
	byteWidth := int(unsafe.Sizeof(elems[0]))
	bitWidth := WidthB(byteWidth)
	if !(WidthU(uint64(l)) <= bitWidth) {
		return 0, ErrSizeOverflow
	}
	if !fixed {
		b.WriteUInt(uint64(l), byteWidth)
	}
	vloc := len(b.buf)
	for i := 0; i < l; i++ {
		b.Write(elems[i], byteWidth)
	}
	fixedLen := 0
	if fixed {
		fixedLen = 0
	}
	b.stack = append(b.stack, value{
		d:           int64(vloc),
		typ:         ToTypedVector(vectorType, fixedLen),
		minBitWidth: bitWidth,
	})
	return vloc, nil
}
