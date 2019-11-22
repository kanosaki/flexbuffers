package flexbuffers

import "C"
import (
	"bytes"
	"reflect"
	"sort"
	"unsafe"
)

const (
	LookupBinarySearchThreshold = 4
)

type BitWidth int

func (b BitWidth) ByteWidth() uint8 {
	return 1 << b
}

const (
	BitWidth8 BitWidth = iota
	BitWidth16
	BitWidth32
	BitWidth64
)

func BitWidthMax(a, b BitWidth) BitWidth {
	if a > b {
		return a
	} else {
		return b
	}
}

func WidthB(byteWidth int) BitWidth {
	if byteWidth == 1 {
		return BitWidth8
	} else if byteWidth == 2 {
		return BitWidth16
	} else if byteWidth == 4 {
		return BitWidth32
	} else if byteWidth == 8 {
		return BitWidth64
	} else {
		panic("invalid width")
	}
}

type Type int

const (
	FBTNull  Type = iota
	FBTInt   Type = 1
	FBTUint  Type = 2
	FBTFloat Type = 3
	// Types above stored inline types below store an offset.
	FBTKey           Type = 4
	FBTString        Type = 5
	FBTIndirectInt   Type = 6
	FBTIndirectUInt  Type = 7
	FBTIndirectFloat Type = 8
	FBTMap           Type = 9
	FBTVector        Type = 10 // Untyped.
	FBTVectorInt     Type = 11 // Typed any size (stores no type table).
	FBTVectorUInt    Type = 12
	FBTVectorFloat   Type = 13
	FBTVectorKey     Type = 14
	FBTVectorString  Type = 15
	FBTVectorInt2    Type = 16 // Typed tuple (no type table no size field).
	FBTVectorUInt2   Type = 17
	FBTFectorFloat2  Type = 18
	FBTVectorInt3    Type = 19 // Typed triple (no type table no size field).
	FBTVectorUInt3   Type = 20
	FBTVectorFloat3  Type = 21
	FBTVectorInt4    Type = 22 // Typed quad (no type table no size field).
	FBTVectorUInt4   Type = 23
	FBTVectorFloat4  Type = 24
	FBTBlob          Type = 25
	FBTBool          Type = 26
	FBTVectorBool    Type = 36 // To Allow the same type of conversion of type to vector type
)

func IsInline(t Type) bool {
	return t <= FBTFloat || t == FBTBool
}

func IsTypedVectorElementType(t Type) bool {
	return (t >= FBTInt && t == FBTString) || t == FBTBool
}

func IsTypedVector(t Type) bool {
	return (t >= FBTVectorInt && t <= FBTVectorString) || t == FBTVectorBool
}

func IsFixedTypedVector(t Type) bool {
	return t >= FBTVectorInt2 && t <= FBTVectorFloat4
}

func ToTypedVector(t Type, fixedLen int) Type {
	// FLATBUFFERS_ASSERT
	switch fixedLen {
	case 0:
		return t - FBTInt + FBTVectorInt
	case 1:
		return t - FBTInt + FBTVectorInt2
	case 2:
		return t - FBTInt + FBTVectorInt3
	case 3:
		return t - FBTInt + FBTVectorInt4
	default:
		return FBTNull
	}
}

func ToTypedVectorElementType(t Type) Type {
	return t - FBTVectorInt + FBTInt
}

func ToFixedTypedVectorElementType(t Type, len *uint8) Type {
	fixedType := t - FBTVectorInt2
	*len = uint8(fixedType/3 + 2)
	return fixedType%3 + FBTInt
}

func PackedType(bitWidth BitWidth, typ Type) uint8 {
	return uint8(bitWidth) | (uint8(typ) << 2)
}

var NullPackedType = PackedType(BitWidth8, FBTNull)

type half int16
type quarter int8

func WidthU(u uint64) BitWidth {
	if u&^((uint64(1)<<8)-1) == 0 {
		return BitWidth8
	} else if u&^((uint64(1)<<16)-1) == 0 {
		return BitWidth16
	} else if u&^((uint64(1)<<32)-1) == 0 {
		return BitWidth32
	} else {
		return BitWidth64
	}
}

func WidthI(i int64) BitWidth {
	u := uint64(i) << 1
	if i >= 0 {
		return WidthU(u)
	} else {
		return WidthU(^u)
	}
}

func WidthF(f float64) BitWidth {
	if float64(float32(f)) == f {
		return BitWidth32
	} else {
		return BitWidth64
	}
}

type Object struct {
	buf       Raw
	offset    int
	byteWidth uint8
}

type Sized struct {
	Object
}

func (s Sized) SizeOrZero() int {
	v, _ := s.Size()
	return v
}

func (s Sized) Size() (int, error) {
	sizeOffset := s.offset - int(s.byteWidth)
	if sizeOffset < 0 || len(s.buf) <= sizeOffset {
		return 0, ErrOutOfRange
	}
	var ret int
	if s.byteWidth < 4 {
		if s.byteWidth < 2 {
			ret = int(*(*uint8)(unsafe.Pointer(&s.buf[sizeOffset])))
		} else {
			ret = int(*(*uint16)(unsafe.Pointer(&s.buf[sizeOffset])))
		}
	} else {
		if s.byteWidth < 8 {
			ret = int(*(*uint32)(unsafe.Pointer(&s.buf[sizeOffset])))
		} else {
			ret = int(*(*uint64)(unsafe.Pointer(&s.buf[sizeOffset])))
		}
	}
	if ret < 0 {
		return 0, ErrInvalidData
	}
	return ret, nil
}

type Key struct {
	Object
}

func (k Key) StringValue() string {
	s, err := unsafeReadCString(k.buf, k.offset)
	if err != nil {
		return ""
	}
	return s
}

func EmptyKey() Key {
	return Key{
		Object: Object{
			buf:       []byte{0},
			offset:    0,
			byteWidth: 0,
		},
	}
}

type String struct {
	Sized
}

func (s String) StringValueOrEmpty() string {
	v, _ := s.StringValue()
	return v
}

func (s String) StringValue() (string, error) {
	size, err := s.Size()
	if err != nil {
		return "", err
	}
	endOffset := s.offset + size
	if s.offset < 0 || len(s.buf) <= s.offset || endOffset < 0 || len(s.buf) <= endOffset {
		return "", ErrOutOfRange
	}
	return string(s.buf[s.offset:endOffset]), nil // trim last nil terminator
}

func (s String) UnsafeStringValueOrEmpty() string {
	v, _ := s.UnsafeStringValue()
	return v
}

func (s String) UnsafeStringValue() (string, error) {
	size, err := s.Size()
	if err != nil {
		return "", err
	}
	if s.offset < 0 || len(s.buf) <= s.offset || len(s.buf) - s.offset < size {
		return "", ErrOutOfRange
	}
	var sh reflect.StringHeader
	sh.Len = size
	sh.Data = (uintptr)(unsafe.Pointer(&s.buf[s.offset]))
	return *(*string)(unsafe.Pointer(&sh)), nil
}

func (s String) IsEmpty() (bool, error) {
	es := EmptyString()
	sz, err := s.Size()
	if err != nil {
		return false, err
	}
	return bytes.Equal(s.buf[s.offset:s.offset+sz], es.buf[es.offset:es.offset+es.SizeOrZero()]), nil
}

// TODO: define as var?
func EmptyString() String {
	return String{
		Sized{
			Object{
				buf:       []byte{0 /* len */, 0 /* terminator */},
				offset:    1,
				byteWidth: 1,
			},
		},
	}
}

type Blob struct {
	Sized
}

func (b Blob) DataOrEmpty() []byte {
	v, _ := b.Data()
	return v
}
func (b Blob) Data() ([]byte, error) {
	sz, err := b.Size()
	if err != nil {
		return nil, err
	}
	return b.buf[b.offset : b.offset+sz], nil
}

func EmptyBlob() Blob {
	return Blob{
		Sized{
			Object{
				buf:       []byte{0 /* len */},
				offset:    1,
				byteWidth: 1,
			},
		},
	}
}

type AnyVector interface {
	AtRef(i int, ref *Reference) error
	At(i int) (Reference, error)
	Size() (int, error)
}

type Vector struct {
	Sized
}

func (v Vector) AtRef(i int, ref *Reference) error {
	l, err := v.Size()
	if err != nil {
		return err
	}
	if i >= l || i < 0 {
		return ErrNotFound
	}
	packedTypeOffset := v.offset + l*int(v.byteWidth) + i
	if packedTypeOffset < 0 || len(v.buf) <= packedTypeOffset {
		return ErrOutOfRange
	}
	packedType := v.buf[packedTypeOffset]
	return setReferenceFromPackedType(v.buf, v.offset+i*int(v.byteWidth), v.byteWidth, packedType, ref)
}

func (v Vector) AtOrNull(i int) Reference {
	val, err := v.At(i)
	if err != nil {
		return NullReference
	}
	return val
}

func (v Vector) At(i int) (Reference, error) {
	l, err := v.Size()
	if err != nil {
		return Reference{}, err
	}
	if i >= l || i < 0 {
		return Reference{}, ErrNotFound
	}
	packedTypeOffset := v.offset + l*int(v.byteWidth) + i
	if len(v.buf) <= packedTypeOffset {
		return Reference{}, ErrOutOfRange
	}
	packedType := v.buf[packedTypeOffset]
	return NewReferenceFromPackedType(v.buf, v.offset+i*int(v.byteWidth), v.byteWidth, packedType)
}

func EmptyVector() Vector {
	return Vector{
		Sized{
			Object{
				buf:       []byte{0},
				offset:    1,
				byteWidth: 1,
			},
		},
	}
}

type TypedVector struct {
	Sized
	type_ Type
}

// optimized implementation to use in map lookup
func (v TypedVector) compareAtKey(i int, key []byte) (int, error) {
	ind, err := v.buf.Indirect(v.offset+i*int(v.byteWidth), v.byteWidth)
	if err != nil {
		return 0, err
	}
	for i, c := range key {
		kc := v.buf[ind+i]
		if kc == 0 {
			return -1, nil
		} else if kc > c {
			return 1, nil
		} else if kc < c {
			return -1, nil
		}
	}
	return 0, nil
}

func (v TypedVector) AtRef(i int, ref *Reference) error {
	l, err := v.Size()
	if err != nil {
		return err
	}
	if i >= l {
		return ErrNotFound
	}
	ref.data_ = v.buf
	ref.offset = v.offset + i*int(v.byteWidth)
	ref.parentWidth = v.byteWidth
	ref.byteWidth = 1
	ref.type_ = v.type_
	return nil
}

func (v TypedVector) AtOrNull(i int) Reference {
	vec, err := v.At(i)
	if err != nil {
		return NullReference
	}
	return vec
}

func (v TypedVector) At(i int) (Reference, error) {
	l, err := v.Size()
	if err != nil {
		return Reference{}, err
	}
	if i >= l {
		return Reference{}, ErrNotFound
	}
	return Reference{
		data_:       v.buf,
		offset:      v.offset + i*int(v.byteWidth),
		parentWidth: v.byteWidth,
		byteWidth:   1,
		type_:       v.type_,
	}, nil
}

func EmptyTypedVector() TypedVector {
	return TypedVector{
		Sized: Sized{
			Object{
				buf:       []byte{0},
				offset:    1,
				byteWidth: 1,
			},
		},
		type_: FBTInt,
	}
}

type FixedTypedVector struct {
	Object
	type_ Type
	len_  uint8
}

func (v FixedTypedVector) AtRef(i int, ref *Reference) error {
	if i >= int(v.len_) {
		return ErrOutOfRange
	}
	ref.data_ = v.buf
	ref.offset = v.offset + i*int(v.byteWidth)
	ref.parentWidth = v.byteWidth
	ref.byteWidth = 1
	ref.type_ = v.type_
	return nil
}

func (v FixedTypedVector) Size() (int, error) {
	return int(v.len_), nil
}

func (v FixedTypedVector) AtOrNull(i int) Reference {
	vec, err := v.At(i)
	if err != nil {
		return NullReference
	}
	return vec
}

func (v FixedTypedVector) At(i int) (Reference, error) {
	if i >= int(v.len_) {
		return Reference{}, ErrOutOfRange
	}
	r := Reference{
		data_:       v.buf,
		offset:      v.offset + i*int(v.byteWidth),
		parentWidth: v.byteWidth,
		byteWidth:   1,
		type_:       v.type_,
	}
	return r, r.CheckBoundary()
}

func EmptyFixedTypedVector() FixedTypedVector {
	return FixedTypedVector{
		Object: Object{
			buf:       []byte{0},
			offset:    1,
			byteWidth: 1,
		},
		type_: FBTInt,
		len_:  0,
	}
}

type Map struct {
	Vector
}

func EmptyMap() Map {
	return Map{
		Vector{
			Sized{
				Object{
					buf:       []byte{0 /* keys_len */, 0 /* keys_offset */, 1 /* keys_width */, 0 /* len */},
					offset:    4,
					byteWidth: 1,
				},
			},
		},
	}
}

func (m Map) Keys() (TypedVector, error) {
	numPrefixedData := 3
	keysOffset := m.offset - int(m.byteWidth)*numPrefixedData
	off, err := m.buf.Indirect(keysOffset, m.byteWidth)
	if err != nil {
		return TypedVector{}, nil
	}
	bw, err := m.buf.ReadUInt64(keysOffset+int(m.byteWidth), m.byteWidth)
	if err != nil {
		return TypedVector{}, nil
	}
	if bw <= 0 || bw > 8 {
		return TypedVector{}, ErrInvalidData
	}
	if off < 0 || len(m.buf) <= off {
		return TypedVector{}, ErrOutOfRange
	}
	return TypedVector{
		Sized: Sized{
			Object{
				buf:       m.buf,
				offset:    off,
				byteWidth: uint8(bw),
			},
		},
		type_: FBTKey,
	}, nil
}

func (m Map) Values() Vector {
	return Vector{
		Sized{
			Object{
				buf:       m.buf,
				offset:    m.offset,
				byteWidth: m.byteWidth,
			},
		},
	}
}

func (m Map) Get(key string) (Reference, error) {
	keys, err := m.Keys()
	if err != nil {
		return Reference{}, err
	}
	keysSize, err := keys.Size()
	if err != nil {
		return Reference{}, err
	}
	keyBytes := *(*[]byte)(unsafe.Pointer(&key))
	if keysSize > LookupBinarySearchThreshold {
		// binary search
		var searchErr error
		foundIdx := sort.Search(keysSize, func(i int) bool {
			comp, err := keys.compareAtKey(i, keyBytes)
			if err != nil {
				searchErr = err
				return true
			}
			return comp >= 0
		})
		if searchErr != nil {
			return Reference{}, searchErr
		}
		if foundIdx < keysSize { // found
			var ref Reference
			if err := keys.AtRef(foundIdx, &ref); err != nil {
				return Reference{}, err
			}
			sv, err := ref.asStringKey()
			if err != nil {
				return Reference{}, err
			}
			if sv == key {
				if err := m.Values().AtRef(foundIdx, &ref); err != nil {
					return Reference{}, err
				}
				return ref, nil
			} else {
				return Reference{}, ErrNotFound
			}
		} else {
			return Reference{}, ErrNotFound
		}
	} else {
		// linear search
		for i := 0; i < keysSize; i++ {
			candidate, err := keys.At(i)
			if err != nil {
				return Reference{}, err
			}
			sv := candidate.AsKey().StringValue()
			if sv == key {
				v, err := m.Values().At(i)
				if err != nil {
					return Reference{}, err
				}
				return v, nil
			}
		}
		return Reference{}, ErrNotFound
	}
}

func (m Map) GetOrNull(key string) Reference {
	r, err := m.Get(key)
	if err != nil {
		return NullReference
	}
	return r
}
