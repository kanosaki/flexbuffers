package flexbuffers

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/cespare/xxhash"
)

var (
	ErrSizeOverflow      = errors.New("overflow")
	ErrOddSizeMapContent = errors.New("map expecting even items, but got odd items")
	ErrOffsetOutOfRange  = errors.New("offset out of range")
	ErrNoNullByte        = errors.New("no null terminator found")
)

type BuilderFlag int

const (
	BuilderFlagNone                BuilderFlag = 0
	BuilderFlagShareKeys           BuilderFlag = 1
	BuilderFlagShareStrings        BuilderFlag = 2
	BuilderFlagShareKeysAndStrings BuilderFlag = 3
	BuilderFlagShareKeyVectors     BuilderFlag = 4
	BuilderFlagShareAll            BuilderFlag = 7
)

type Builder struct {
	buf              []byte
	stack            []value
	finished         bool
	flags            BuilderFlag
	forceMinBitWidth BitWidth
	err              error

	keyOffsetMap        map[uint64]uint64
	stringOffsetMap     map[uint64]uint64
	keyVectorsOffsetMap map[uint64]uint64
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
	if b.flags&BuilderFlagShareKeys == BuilderFlagShareKeys {
		b.keyOffsetMap = make(map[uint64]uint64)
	} else {
		b.keyOffsetMap = nil
	}
	if b.flags&BuilderFlagShareStrings == BuilderFlagShareStrings {
		b.stringOffsetMap = make(map[uint64]uint64)
	} else {
		b.stringOffsetMap = nil
	}
	if b.flags&BuilderFlagShareKeyVectors == BuilderFlagShareKeyVectors {
		b.keyVectorsOffsetMap = make(map[uint64]uint64)
	} else {
		b.keyVectorsOffsetMap = nil
	}
}

func (b *Builder) Finish() error {
	if b.err != nil {
		return b.err
	}
	if len(b.stack) == 0 {
		return fmt.Errorf("empty document")
	}
	byteWidth := b.align(b.stack[0].ElemWidth(len(b.buf), 0))
	if err := b.WriteAny(&b.stack[0], byteWidth); err != nil {
		return err
	}
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
	if b.flags&BuilderFlagShareKeys == BuilderFlagShareKeys {
		hash := xxhash.Sum64(key)
		if prevSloc, ok := b.keyOffsetMap[hash]; ok {
			prevKey := readCStringBytes(b.buf, int(prevSloc))
			if bytes.Compare(key, prevKey) == 0 {
				b.stack = append(b.stack, newValueUInt(prevSloc, FBTKey, BitWidth8))
				return prevSloc
			}
		}
		b.keyOffsetMap[hash] = sloc
	}
	b.WriteBytes(key)
	b.buf = append(b.buf, 0) // terminate bytes
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

func (b *Builder) IndirectIntField(key []byte, i int64) {
	b.Key(key)
	b.IndirectInt(i)
}

func (b *Builder) IndirectUIntField(key []byte, i uint64) {
	b.Key(key)
	b.IndirectUInt(i)
}

func (b *Builder) IndirectFloat32Field(key []byte, f float32) {
	b.Key(key)
	b.IndirectFloat32(f)
}

func (b *Builder) IndirectFloat64Field(key []byte, f float64) {
	b.Key(key)
	b.IndirectFloat64(f)
}

func (b *Builder) StringValue(s string) int {
	//resetTo := b.buf.Len()
	sloc := b.createBlob(stringToBytes(s), 1, FBTString)
	if b.flags&BuilderFlagShareStrings == BuilderFlagShareStrings {
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
	if b.err != nil {
		return 0
	}
	start := b.StartMapField(key)
	fn(b)
	idx, err := b.EndMap(start)
	if err != nil {
		b.err = err
	}
	return idx
}

func (b *Builder) Map(fn func(bld *Builder)) int {
	if b.err != nil {
		return 0
	}
	start := b.StartMap()
	fn(b)
	idx, err := b.EndMap(start)
	if err != nil {
		b.err = err
	}
	return idx
}

func (b *Builder) StartMapField(key []byte) int {
	b.Key(key)
	return len(b.stack)
}

func (b *Builder) EndVector(start int, typed, fixed bool) (uint64, error) {
	vec, err := b.createVector(start, len(b.stack)-start, 1, typed, fixed, nil)
	if err != nil {
		return 0, err
	}
	b.stack = b.stack[:start]
	b.stack = append(b.stack, vec)
	return vec.AsUInt(), nil
}

func (b *Builder) Vector(typed, fixed bool, fn func(bld *Builder)) int {
	if b.err != nil {
		return 0
	}
	start := b.StartVector()
	fn(b)
	off, err := b.EndVector(start, typed, fixed)
	if err != nil {
		b.err = err
	}
	return int(off)
}
func (b *Builder) VectorField(key []byte, typed, fixed bool, fn func(bld *Builder)) int {
	if b.err != nil {
		return 0
	}
	b.Key(key)
	return b.Vector(typed, fixed, fn)
}

func (b *Builder) EndMap(start int) (int, error) {
	l := len(b.stack) - start
	if l&1 > 0 {
		return 0, ErrOddSizeMapContent
	}
	l /= 2
	for key := start; key < len(b.stack); key += 2 {
		if b.stack[key].typ != FBTKey {
			return 0, fmt.Errorf("odd elemnt of map must be Key")
		}
	}
	// TODO: implement sorting
	keys, err := b.createVector(start, l, 2, true, false, nil)
	if err != nil {
		return 0, err
	}
	vec, err := b.createVector(start+1, l, 2, false, false, &keys)
	if err != nil {
		return 0, err
	}
	b.stack = b.stack[:start]
	b.stack = append(b.stack, vec)
	return int(vec.AsUInt()), nil
}

func (b *Builder) WriteOffset(o int, byteWidth int) error {
	reloff := len(b.buf) - o
	if byteWidth != 8 && reloff >= 1<<(byteWidth*8) {
		return ErrOffsetOutOfRange
	}
	b.WriteUInt(uint64(reloff), byteWidth)
	return nil
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

func (b *Builder) WriteAny(v *value, byteWidth int) error {
	switch v.typ {
	case FBTNull, FBTInt:
		b.WriteInt(v.AsInt(), byteWidth)
	case FBTBool, FBTUint:
		b.WriteUInt(v.AsUInt(), byteWidth)
	case FBTFloat:
		return b.WriteDouble(v.AsFloat(), byteWidth)
	default:
		return b.WriteOffset(int(v.AsUInt()), byteWidth)
	}
	return nil
}

func (b *Builder) createVector(start, vecLen, step int, typed, fixed bool, keys *value) (value, error) {
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
				return value{}, fmt.Errorf("inconsistent type")
			}
		}
	}
	if fixed && !IsTypedVectorElementType(vectorType) {
		return value{}, fmt.Errorf("item type should be one of Int / UInt / Float / Key")
	}
	byteWidth := b.align(bitWidth)
	if keys != nil {
		if err := b.WriteOffset(int(keys.d), byteWidth); err != nil {
			return value{}, err
		}
		b.WriteUInt(1<<keys.minBitWidth, byteWidth)
	}
	if !fixed {
		b.WriteUInt(uint64(vecLen), byteWidth)
	}
	vloc := len(b.buf)
	for i := start; i < len(b.stack); i += step {
		if err := b.WriteAny(&b.stack[i], byteWidth); err != nil {
			return value{}, err
		}
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
	}, nil
}

func (b *Builder) Write(v interface{}, byteWidth int) {

}

func GetScalarType(v interface{}) (Type, error) {
	switch v.(type) {
	case float32, float64:
		return FBTFloat, nil
	case int8, int16, int32, int64:
		return FBTInt, nil
	case uint8, uint16, uint32, uint64:
		return FBTUint, nil
	default:
		return FBTNull, fmt.Errorf("type %s is not scalar type", reflect.TypeOf(v).Name())
	}
}
