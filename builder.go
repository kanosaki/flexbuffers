package flexbuffers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/cespare/xxhash"
)

var (
	ErrSizeOverflow      = errors.New("overflow")
	ErrOddSizeMapContent = errors.New("map expecting even items, but got odd items")
	ErrOutOfRange        = errors.New("out of range, might be broken data")
	ErrInvalidData       = errors.New("invalid data")
	ErrRecursiveData     = errors.New("invalid data: data has a cyclic offset")
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

type offsetAndLen struct {
	offset uint64
	size   int
}

type Builder struct {
	buf              []byte
	stack            []value
	finished         bool
	flags            BuilderFlag
	forceMinBitWidth BitWidth
	err              error
	ext              int64
	extMap           map[int]int64

	keyOffsetMap        map[uint64]offsetAndLen
	stringOffsetMap     map[uint64]offsetAndLen
	keyVectorsOffsetMap map[uint64]value
}

func NewBuilder() *Builder {
	b := new(Builder)
	b.Clear()
	return b
}

func NewBuilderWithFlags(flags BuilderFlag) *Builder {
	b := new(Builder)
	b.flags = flags
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
	b.extMap = make(map[int]int64)
	if b.flags&BuilderFlagShareKeys == BuilderFlagShareKeys {
		b.keyOffsetMap = make(map[uint64]offsetAndLen)
	} else {
		b.keyOffsetMap = nil
	}
	if b.flags&BuilderFlagShareStrings == BuilderFlagShareStrings {
		b.stringOffsetMap = make(map[uint64]offsetAndLen)
	} else {
		b.stringOffsetMap = nil
	}
	if b.flags&BuilderFlagShareKeyVectors == BuilderFlagShareKeyVectors {
		b.keyVectorsOffsetMap = make(map[uint64]value)
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

func (b *Builder) Ext(i int64) {
	b.ext = i
}

func (b *Builder) WriteBytes(data []byte) {
	b.buf = append(b.buf, data...)
}

func (b *Builder) Key(key []byte) uint64 {
	sloc := uint64(len(b.buf))
	var conflict bool
	var hash uint64
	share := b.flags&BuilderFlagShareKeys == BuilderFlagShareKeys
	if share {
		hash = xxhash.Sum64(key)
		prev, ok := b.keyOffsetMap[hash]
		if ok && bytes.Compare(key, b.buf[prev.offset:int(prev.offset)+prev.size]) == 0 {
			sloc = prev.offset
			b.stack = append(b.stack, newValueUInt(prev.offset, FBTKey, BitWidth8, false))
			return prev.offset
		}
		// not found or found but different content (hash conflict)
		conflict = ok
	}
	b.WriteBytes(key)
	b.buf = append(b.buf, 0) // terminate bytes
	b.stack = append(b.stack, newValueUInt(sloc, FBTKey, BitWidth8, false))
	if share && !conflict {
		b.keyOffsetMap[hash] = offsetAndLen{
			offset: sloc,
			size:   len(key),
		}
	}
	return sloc
}

func (b *Builder) AttachMetadata(tag int, body []byte) {

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
	b.stack = append(b.stack, newValueUInt(i, FBTUint, WidthU(i), false))
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
	var hash uint64
	var conflict bool
	share := b.flags&BuilderFlagShareStrings == BuilderFlagShareStrings
	data := stringToBytes(s)
	if share {
		hash = xxhash.Sum64String(s)
		prevLoc, ok := b.stringOffsetMap[hash]
		if ok && bytes.Compare(data, b.buf[prevLoc.offset:int(prevLoc.offset)+prevLoc.size]) == 0 {
			bitWidth := WidthU(uint64(len(data)))
			b.stack = append(b.stack, newValueUInt(prevLoc.offset, FBTString, bitWidth, b.ext != 0))
			return int(prevLoc.offset)
		}
		// not found or found but different content (hash conflict)
		conflict = ok
	}
	sloc := b.createBlob(data, 1, FBTString)
	if share && !conflict {
		b.stringOffsetMap[hash] = offsetAndLen{
			offset: uint64(sloc),
			size:   len(data),
		}
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
	n := len(b.stack)
	if b.ext != 0 {
		b.extMap[n] = b.ext
	}
	return n
}

func (b *Builder) StartVectorField(key []byte) int {
	b.Key(key)
	return len(b.stack)
}

func (b *Builder) StartMap() int {
	n := len(b.stack)
	if b.ext != 0 {
		b.extMap[n] = b.ext
	}
	return n
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
	ext := b.extMap[start]
	vec, err := b.createVector(start, len(b.stack)-start, 1, typed, fixed, nil, ext, ext != 0)
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

type keysValueSlice struct {
	b      *Builder
	values []value
}

func (k *keysValueSlice) Len() int {
	return len(k.values) / 2
}

func (k *keysValueSlice) Less(i, j int) bool {
	iKey := readCStringBytes(k.b.buf, int(k.values[2*i].d))
	jKey := readCStringBytes(k.b.buf, int(k.values[2*j].d))
	return bytes.Compare(iKey, jKey) <= 0
}

func (k *keysValueSlice) Swap(i, j int) {
	k.values[2*i], k.values[2*j] = k.values[2*j], k.values[2*i]
	k.values[2*i+1], k.values[2*j+1] = k.values[2*j+1], k.values[2*i+1]
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
	sortingSlice := keysValueSlice{
		b:      b,
		values: b.stack[start:],
	}
	sort.Sort(&sortingSlice)

	ext := b.extMap[start]
	share := b.flags&BuilderFlagShareKeyVectors == BuilderFlagShareKeyVectors
	var keys value
	var err error
	var hash uint64
	if share && ext == 0 {
		h := xxhash.New()
		for key := start; key < len(b.stack); key += 2 {
			_, _ = h.Write(readCStringBytes(b.buf, int(b.stack[key].d)))
		}
		hash = h.Sum64()
		if prev, ok := b.keyVectorsOffsetMap[hash]; ok {
			keys = prev
			goto keyOk
		}
	}
	// attach ext only after keys vector
	keys, err = b.createVector(start, l, 2, true, false, nil, ext, false)
	if err != nil {
		return 0, err
	}
	if share {
		b.keyVectorsOffsetMap[hash] = keys
	}
keyOk:
	vec, err := b.createVector(start+1, l, 2, false, false, &keys, 0, ext != 0)
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
		return ErrOutOfRange
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
	hasExt := b.ext != 0
	if hasExt {
		var buf [8]byte
		l := binary.PutVarint(buf[:], b.ext)
		b.buf = append(b.buf, buf[:l]...)
		b.ext = 0
	}
	b.stack = append(b.stack, newValueUInt(uint64(sloc), t, bitWidth, hasExt))
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

func (b *Builder) createVector(start, vecLen, step int, typed, fixed bool, keys *value, ext int64, hasExtType bool) (value, error) {
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
	hasExt := ext != 0
	if hasExt {
		var buf [8]byte
		l := binary.PutVarint(buf[:], ext)
		b.buf = append(b.buf, buf[:l]...)
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
		hasExt:      hasExtType,
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
