//+build !unsafe

package flexbuffers

import (
	"encoding/binary"
	"math"
	"unsafe"
)

func (b *Builder) IndirectInt(i int64) {
	var tmp [8]byte
	bitWidth := WidthI(i)
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	binary.LittleEndian.PutUint64(tmp[:], uint64(i))
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectInt, bitWidth))
}

func (b *Builder) IndirectUInt(i uint64) {
	var tmp [8]byte
	bitWidth := WidthU(i)
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	binary.LittleEndian.PutUint64(tmp[:], i)
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectUInt, bitWidth))
}

func (b *Builder) IndirectFloat32(f float32) {
	var tmp [4]byte
	bitWidth := BitWidth32
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	binary.LittleEndian.PutUint64(tmp[:], uint64(math.Float32bits(f)))
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectFloat, bitWidth))
}

func (b *Builder) IndirectFloat64(f float64) {
	var tmp [8]byte
	bitWidth := WidthF(f)
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	binary.LittleEndian.PutUint64(tmp[:], math.Float64bits(f))
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectFloat, bitWidth))
}

func (b *Builder) allocate(bw int) int {
	l := len(b.buf)
	if bw <= cap(b.buf)-l {
		b.buf = b.buf[:l+bw]
	} else {
		b.buf = append(b.buf, make([]byte, bw)...)
	}
	return l
}

func (b *Builder) WriteUInt(i uint64, bw int) {
	offset := b.allocate(bw)
	if bw == 1 {
		b.buf[offset] = uint8(i)
	} else if bw == 2 {
		binary.LittleEndian.PutUint16(b.buf[offset:], uint16(i))
	} else if bw == 4 {
		binary.LittleEndian.PutUint32(b.buf[offset:], uint32(i))
	} else if bw == 8 {
		binary.LittleEndian.PutUint64(b.buf[offset:], uint64(i))
	}
}

func (b *Builder) WriteInt(i int64, bw int) {
	offset := b.allocate(bw)
	if bw == 1 {
		b.buf[offset] = uint8(i)
	} else if bw == 2 {
		binary.LittleEndian.PutUint16(b.buf[offset:], uint16(i))
	} else if bw == 4 {
		binary.LittleEndian.PutUint32(b.buf[offset:], uint32(i))
	} else if bw == 8 {
		binary.LittleEndian.PutUint64(b.buf[offset:], uint64(i))
	}
}
func (b *Builder) WriteDouble(f float64, bw int) error {
	offset := b.allocate(bw)
	if bw == 4 {
		binary.LittleEndian.PutUint32(b.buf[offset:], math.Float32bits(float32(f)))
	} else if bw == 8 {
		binary.LittleEndian.PutUint64(b.buf[offset:], math.Float64bits(f))
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
