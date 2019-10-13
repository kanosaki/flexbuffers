//+build amd64

package flexbuffers

import (
	"fmt"
	"unsafe"
)

type BuilderFlag int

const (
	BuilderFlagNone BuilderFlag = iota
	BuilderFlagShareKeys
	BuilderFlagShareStrings
	BuilderFlagShareKeysAndStrings
	BuilderFlagShareKeyVectors
	BuilderFlagShareAll
)

type stringOffset [2]string

type Builder struct {
	buf              []byte
	stack            []value
	finished         bool
	flags            BuilderFlag
	forceMinBitWidth BitWidth

	keyOffsetMap    map[int]struct{}
	stringOffsetMap map[stringOffset]struct{}
}

func NewBuilder() *Builder {
	b := new(Builder)
	b.Clear()
	return b
}

func (b *Builder) Buffer() Raw {
	if !b.finished {
		panic("assertion error")
	}
	return b.buf
}

func (b *Builder) Size() int {
	return len(b.buf)
}

func (b *Builder) Clear() {
	b.buf = make([]byte, 0, 64)
	b.stack = nil
	b.finished = false
	b.forceMinBitWidth = BitWidth8
	b.keyOffsetMap = make(map[int]struct{})
	b.stringOffsetMap = make(map[stringOffset]struct{})
}

func (b *Builder) Finish() error {
	if len(b.stack) == 0 {
		return fmt.Errorf("empty document")
	}
	byteWidth := b.align(b.stack[0].ElemWidth(len(b.buf), 0))
	b.WriteAny(&b.stack[0], byteWidth)
	b.WriteUInt(uint64(b.stack[0].StoredPackedType(BitWidth8)), 1)
	b.WriteUInt(uint64(byteWidth), 1)
	b.finished = true
	return nil
}

func (b *Builder) WriteBytes(data []byte) {
	b.buf = append(b.buf, data...)
}

func (b *Builder) Key(key []byte) uint64 {
	sloc := uint64(len(b.buf))
	b.WriteBytes(key)
	b.buf = append(b.buf, 0) // terminate bytes
	if b.flags&BuilderFlagShareStrings > 0 {
		// TODO
	}
	b.stack = append(b.stack, newValueUInt(sloc, FBTKey, BitWidth8))
	return sloc
}

func (b *Builder) Null() {
	b.stack = append(b.stack, value{})
}

func (b *Builder) NullField(key []byte) {
	b.Key(key)
	b.Null()
}

func (b *Builder) Int(i int64) {
	b.stack = append(b.stack, newValueInt(i, FBTInt, WidthI(i)))
}

func (b *Builder) IntField(key []byte, i int64) {
	b.Key(key)
	b.Int(i)
}

func (b *Builder) UInt(i uint64) {
	b.stack = append(b.stack, newValueUInt(i, FBTUint, WidthU(i)))
}

func (b *Builder) UIntField(key []byte, i uint64) {
	b.Key(key)
	b.UInt(i)
}

func (b *Builder) Float32(f float32) {
	b.stack = append(b.stack, newValueFloat32(f))
}

func (b *Builder) Float32Field(key []byte, f float32) {
	b.Key(key)
	b.Float32(f)
}

func (b *Builder) Float64(f float64) {
	b.stack = append(b.stack, newValueFloat64(f))
}

func (b *Builder) Float64Field(key []byte, f float64) {
	b.Key(key)
	b.Float64(f)
}

func (b *Builder) Bool(v bool) {
	b.stack = append(b.stack, newValueBool(v))
}

func (b *Builder) BoolField(key []byte, v bool) {
	b.Key(key)
	b.Bool(v)
}

func (b *Builder) align(alignment BitWidth) int {
	byteWidth := 1 << alignment
	for i := 0; i < PaddingBytes(len(b.buf), byteWidth); i++ {
		b.buf = append(b.buf, 0)
	}
	return byteWidth
}

func (b *Builder) IndirectInt(i int64) {
	var tmp [8]byte
	bitWidth := WidthI(i)
	byteWidth := b.align(bitWidth)
	iloc := uint64(len(b.buf))
	*((*int64)(unsafe.Pointer(&tmp[0]))) = i
	b.WriteBytes(tmp[:byteWidth])
	b.stack = append(b.stack, newValueUInt(iloc, FBTIndirectInt, bitWidth))
}

func (b *Builder) IndirectIntField(key []byte, i int64) {
	b.Key(key)
	b.IndirectInt(i)
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

func (b *Builder) IndirectUIntField(key []byte, i uint64) {
	b.Key(key)
	b.IndirectUInt(i)
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

func (b *Builder) IndirectFloat32Field(key []byte, f float32) {
	b.Key(key)
	b.IndirectFloat32(f)
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

func (b *Builder) IndirectFloat64Field(key []byte, f float64) {
	b.Key(key)
	b.IndirectFloat64(f)
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
func (b *Builder) WriteDouble(f float64, byteWidth int) {
	ptr := b.allocateUnsafe(byteWidth)
	if byteWidth == 4 {
		*(*float64)(ptr) = f
	} else if byteWidth == 8 {
		*(*float32)(ptr) = float32(f)
	} else {
		panic("invalid float width")
	}
}

func (b *Builder) StringValue(s string) int {
	//resetTo := b.buf.Len()
	sloc := b.createBlob(stringToBytes(s), 1, FBTString)
	if b.flags&BuilderFlagShareStrings > 0 {
		// TODO
	}
	return sloc
}

func (b *Builder) StringValueField(key []byte, s string) int {
	b.Key(key)
	b.StringValue(s)
	return len(b.stack)
}

func (b *Builder) Blob(data []byte) int {
	return b.createBlob(data, 0, FBTBlob)
}

func (b *Builder) BlobField(key, data []byte) int {
	b.Key(key)
	b.Blob(data)
	return len(b.stack)
}

func (b *Builder) StartVector() int {
	return len(b.stack)
}

func (b *Builder) StartVectorField(key []byte) int {
	b.Key(key)
	return len(b.stack)
}

func (b *Builder) StartMap() int {
	return len(b.stack)
}

func (b *Builder) MapField(key []byte, fn func(bld *Builder)) int {
	start := b.StartMapField(key)
	fn(b)
	return b.EndMap(start)
}

func (b *Builder) Map(fn func(bld *Builder)) int {
	start := b.StartMap()
	fn(b)
	return b.EndMap(start)
}

func (b *Builder) StartMapField(key []byte) int {
	b.Key(key)
	return len(b.stack)
}

func (b *Builder) EndVector(start int, typed, fixed bool) uint64 {
	vec := b.createVector(start, len(b.stack)-start, 1, typed, fixed, nil)
	b.stack = b.stack[:start]
	b.stack = append(b.stack, vec)
	return vec.AsUInt()
}

func (b *Builder) Vector(typed, fixed bool, fn func(bld *Builder)) int {
	start := b.StartVector()
	fn(b)
	return int(b.EndVector(start, typed, fixed))
}
func (b *Builder) VectorField(key []byte, typed, fixed bool, fn func(bld *Builder)) int {
	b.Key(key)
	return b.Vector(typed, fixed, fn)
}

func (b *Builder) EndMap(start int) int {
	l := len(b.stack) - start
	if l&1 > 0 {
		panic("assertion failed")
	}
	l /= 2
	for key := start; key < len(b.stack); key += 2 {
		if b.stack[key].typ != FBTKey {
			panic("error: all key must be FBTKey")
		}
	}
	// TODO: implement sorting
	keys := b.createVector(start, l, 2, true, false, nil)
	vec := b.createVector(start+1, l, 2, false, false, &keys)
	b.stack = b.stack[:start]
	b.stack = append(b.stack, vec)
	return int(vec.AsUInt())
}

func (b *Builder) WriteOffset(o int, byteWidth int) {
	reloff := len(b.buf) - o
	if byteWidth != 8 && reloff >= 1<<(byteWidth*8) {
		panic("assertion failed")
	}
	b.WriteUInt(uint64(reloff), byteWidth)
}

func (b *Builder) createBlob(data []byte, trailing int, t Type) int {
	bitWidth := WidthU(uint64(len(data)))
	byteWidth := b.align(bitWidth)
	b.WriteUInt(uint64(len(data)), byteWidth)
	sloc := len(b.buf)
	b.WriteBytes(data)
	for i := 0; i < trailing; i++ {
		b.buf = append(b.buf, 0)
	}
	b.stack = append(b.stack, newValueUInt(uint64(sloc), t, bitWidth))
	return sloc
}

func (b *Builder) WriteAny(v *value, byteWidth int) {
	switch v.typ {
	case FBTNull, FBTInt:
		b.WriteInt(v.AsInt(), byteWidth)
	case FBTBool, FBTUint:
		b.WriteUInt(v.AsUInt(), byteWidth)
	case FBTFloat:
		b.WriteDouble(v.AsFloat(), byteWidth)
	default:
		b.WriteOffset(int(v.AsUInt()), byteWidth)
	}
}

func (b *Builder) createVector(start, vecLen, step int, typed, fixed bool, keys *value) value {
	bitWidth := BitWidthMax(b.forceMinBitWidth, WidthU(uint64(vecLen)))
	prefixElems := 1
	if keys != nil {
		bitWidth = BitWidthMax(bitWidth, keys.ElemWidth(len(b.buf), 0))
		prefixElems += 2
	}
	vectorType := FBTKey
	for i := start; i < len(b.stack); i += step {
		elemWidth := b.stack[i].ElemWidth(len(b.buf), i+prefixElems)
		bitWidth = BitWidthMax(bitWidth, elemWidth)
		if typed {
			if i == start {
				vectorType = b.stack[i].typ
			} else if b.stack[i].typ != vectorType {
				panic("inconsistent type")
			}
		}
	}
	if fixed && !IsTypedVectorElementType(vectorType) {
		panic("item type should be one of Int / UInt / Float / Key")
	}
	byteWidth := b.align(bitWidth)
	if keys != nil {
		b.WriteOffset(int(keys.d), byteWidth) // uint64
		b.WriteUInt(1<<keys.minBitWidth, byteWidth)
	}
	if !fixed {
		b.WriteUInt(uint64(vecLen), byteWidth)
	}
	vloc := len(b.buf)
	for i := start; i < len(b.stack); i += step {
		b.WriteAny(&b.stack[i], byteWidth)
	}
	if !typed {
		for i := start; i < len(b.stack); i += step {
			// TODO: optimize by preallocate
			b.buf = append(b.buf, b.stack[i].StoredPackedType(bitWidth))
		}
	}
	t := FBTVector
	if keys != nil {
		t = FBTMap
		if typed {
			if fixed {
				t = ToTypedVector(vectorType, vecLen)
			} else {
				t = ToTypedVector(vectorType, 0)
			}
		}
	}
	return value{
		d:           int64(vloc),
		typ:         t,
		minBitWidth: bitWidth,
	}
}

func (b *Builder) Write(v interface{}, byteWidth int) {

}

func GetScalarType(v interface{}) Type {
	switch v.(type) {
	case float32, float64:
		return FBTFloat
	case int8, int16, int32, int64:
		return FBTInt
	case uint8, uint16, uint32, uint64:
		return FBTUint
	default:
		panic("assertion failed")
	}
}

func (b *Builder) scalarVector(elems []interface{}, fixed bool) int {
	vectorType := GetScalarType(elems[0])
	l := len(elems)
	byteWidth := int(unsafe.Sizeof(elems[0]))
	bitWidth := WidthB(byteWidth)
	if !(WidthU(uint64(l)) <= bitWidth) {
		panic("overflow")
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
	return vloc
}
