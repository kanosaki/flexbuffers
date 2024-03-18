package flexbuffers

import "C"
import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unsafe"

	"github.com/kanosaki/flexbuffers/pkg/unsafeutil"
)

var (
	ErrTypeDoesNotMatch = errors.New("type doesn't match")
)

var NullReference = Reference{
	data_:       nil,
	offset:      0,
	parentWidth: 0,
	byteWidth:   0,
	type_:       FBTNull,
}

type Reference struct {
	data_       Raw
	offset      int
	type_       Type
	parentWidth uint8
	byteWidth   uint8
	hasExt      bool
}

func (r Reference) CheckBoundary() error {
	if r.offset < 0 || len(r.data_) <= r.offset {
		return ErrOutOfRange
	}
	return nil
}

func (r Reference) String() string {
	var sb bytes.Buffer
	if err := r.WriteAsJson(&sb); err != nil {
		return err.Error()
	}
	return sb.String()
}

func (r Reference) writeJsonVector(l int, vec AnyVector, w io.Writer) (err error) {
	if _, err = fmt.Fprintf(w, "["); err != nil {
		return
	}
	for i := 0; i < l; i++ {
		item, err := vec.At(i)
		if err != nil {
			return err
		}
		if err := item.WriteAsJson(w); err != nil {
			return err
		}
		if i != l-1 {
			if _, err := fmt.Fprintf(w, ","); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintf(w, "]"); err != nil {
		return err
	}
	return
}

func (r Reference) WriteAsJson(w io.Writer) (err error) {
	switch r.type_ {
	case FBTNull:
		_, err = fmt.Fprintf(w, "null")
	case FBTInt, FBTIndirectInt:
		i, err := r.Int64()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "%d", i)
	case FBTUint, FBTIndirectUInt:
		i, err := r.UInt64()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "%d", i)
	case FBTFloat, FBTIndirectFloat:
		f, err := r.Float64()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "%f", f)
	case FBTKey:
		k, err := r.asStringKey()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\"%s\"", k)
	case FBTString:
		sRef, err := r.StringRef()
		if err != nil {
			return err
		}
		unsafeStr, err := sRef.StringValue()
		if err != nil {
			return err
		}
		out := make([]byte, 0, len(unsafeStr))
		out = EscapeJSONString(out, unsafeStr)

		_, err = fmt.Fprintf(w, unsafeutil.B2S(out))
	case FBTMap:
		if _, err := fmt.Fprintf(w, "{"); err != nil {
			return err
		}
		m, err := r.Map()
		if err != nil {
			return err
		}
		var keys TypedVector
		keys, err = m.Keys()
		if err != nil {
			return err
		}
		values := m.Values()
		sz, err := keys.Size()
		if err != nil {
			return err
		}
		for i := 0; i < sz; i++ {
			k, err := keys.At(i)
			if err != nil {
				return err
			}
			v, err := values.At(i)
			if err != nil {
				return err
			}
			if err := k.WriteAsJson(w); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, ":"); err != nil {
				return err
			}
			if err := v.WriteAsJson(w); err != nil {
				return err
			}
			if i != sz-1 {
				if _, err := fmt.Fprintf(w, ","); err != nil {
					return err
				}
			}
		}
		if _, err := fmt.Fprintf(w, "}"); err != nil {
			return err
		}
	case FBTVector:
		vec, err := r.Vector()
		if err != nil {
			return err
		}
		sz, err := vec.Size()
		if err != nil {
			return err
		}
		err = r.writeJsonVector(sz, vec, w)
	case FBTBlob:
		blob, err := r.Blob()
		if err != nil {
			return err
		}
		d, err := blob.Data()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\"%s\"", base64.StdEncoding.EncodeToString(d))
	case FBTBool:
		b, err := r.Bool()
		if err != nil {
			return err
		}
		if b {
			_, err = fmt.Fprintf(w, "true")
		} else {
			_, err = fmt.Fprintf(w, "false")
		}
	default:
		if r.IsTypedVector() {
			vec, err := r.TypedVector()
			if err != nil {
				return err
			}
			sz, err := vec.Size()
			if err != nil {
				return err
			}
			err = r.writeJsonVector(sz, vec, w)
		} else if r.IsFixedTypedVector() {
			vec, err := r.FixedTypedVector()
			if err != nil {
				return err
			}
			err = r.writeJsonVector(int(vec.len_), vec, w)
		} else {
			return fmt.Errorf("unable to convert to json: type=%v", r.type_)
		}
	}
	return
}

func setReferenceFromPackedType(buf Raw, offset int, parentWidth uint8, packedType uint8, ref *Reference) error {
	bw, t, hasExt := UnpackType(packedType)
	if offset < 0 || len(buf) <= offset+int(bw) {
		return ErrOutOfRange
	}
	ref.data_ = buf
	ref.offset = offset
	ref.parentWidth = parentWidth
	ref.byteWidth = bw.ByteWidth()
	ref.type_ = t
	ref.hasExt = hasExt
	return nil
}

func NewReferenceFromPackedType(buf Raw, offset int, parentWidth uint8, packedType uint8) (Reference, error) {
	bw, t, hasExt := UnpackType(packedType)
	r := Reference{
		data_:       buf,
		offset:      offset,
		parentWidth: parentWidth,
		byteWidth:   bw.ByteWidth(),
		type_:       t,
		hasExt:      hasExt,
	}
	return r, r.CheckBoundary()
}

func (r Reference) IsNull() bool {
	return r.type_ == FBTNull
}

func (r Reference) IsBool() bool {
	return r.type_ == FBTBool
}

func (r Reference) IsInt() bool {
	return r.type_ == FBTInt || r.type_ == FBTIndirectInt
}

func (r Reference) IsUInt() bool {
	return r.type_ == FBTUint || r.type_ == FBTIndirectUInt
}

func (r Reference) IsIntOrUInt() bool {
	return r.IsInt() || r.IsUInt()
}

func (r Reference) IsFloat() bool {
	return r.type_ == FBTFloat || r.type_ == FBTIndirectFloat
}

func (r Reference) IsNumeric() bool {
	return r.IsIntOrUInt() || r.IsFloat()
}

func (r Reference) IsString() bool {
	return r.type_ == FBTString
}

func (r Reference) IsKey() bool {
	return r.type_ == FBTKey
}

func (r Reference) IsVector() bool {
	return r.type_ == FBTVector || r.type_ == FBTMap
}

func (r Reference) IsUntypedVector() bool {
	return r.type_ == FBTVector
}

func (r Reference) IsTypedVector() bool {
	return IsTypedVector(r.type_)
}

func (r Reference) IsFixedTypedVector() bool {
	return IsFixedTypedVector(r.type_)
}

func (r Reference) IsAnyVector() bool {
	return r.IsTypedVector() || r.IsFixedTypedVector() || r.IsVector()
}

func (r Reference) IsMap() bool {
	return r.type_ == FBTMap
}

func (r Reference) IsBlob() bool {
	return r.type_ == FBTBlob
}
func (r Reference) AsBool() bool {
	v, _ := r.Bool()
	return v
}

func (r Reference) Bool() (bool, error) {
	if r.type_ == FBTBool {
		v, err := r.data_.ReadUInt64(r.offset, r.parentWidth)
		if err != nil {
			return false, err
		}
		return v != 0, nil
	} else {
		v, err := r.UInt64()
		if err != nil {
			return false, err
		}
		return v != 0, nil
	}
}

func (r Reference) indirect() (int, error) {
	return r.data_.Indirect(r.offset, r.parentWidth)
}

func (r Reference) Int64() (int64, error) {
	if r.type_ == FBTInt {
		return r.data_.ReadInt64(r.offset, r.parentWidth)
	} else {
		switch r.type_ {
		case FBTIndirectInt:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			return r.data_.ReadInt64(ind, r.byteWidth)
		case FBTUint:
			v, err := r.data_.ReadUInt64(r.offset, r.parentWidth)
			if err != nil {
				return 0, err
			}
			return int64(v), nil
		case FBTIndirectUInt:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			v, err := r.data_.ReadUInt64(ind, r.byteWidth)
			if err != nil {
				return 0, err
			}
			return int64(v), nil
		case FBTFloat:
			v, err := r.data_.ReadDouble(r.offset, r.parentWidth)
			if err != nil {
				return 0, err
			}
			return int64(v), nil
		case FBTIndirectFloat:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			v, err := r.data_.ReadDouble(ind, r.byteWidth)
			if err != nil {
				return 0, err
			}
			return int64(v), nil
		case FBTNull:
			return 0, nil
		case FBTString:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			s, err := cstringBytesToString(r.data_[ind:])
			if err != nil {
				return 0, err
			}
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return 0, err
			}
			return i, nil
		case FBTVector:
			vec, err := r.Vector()
			if err != nil {
				return 0, err
			}
			sz, err := vec.Size()
			if err != nil {
				return 0, err
			}
			return int64(sz), nil
		case FBTBool:
			return r.data_.ReadInt64(r.offset, r.parentWidth)
		default:
			return 0, ErrTypeDoesNotMatch
		}
	}
}

func (r Reference) AsInt64() int64 {
	v, _ := r.Int64()
	return v
}
func (r Reference) AsUInt64() uint64 {
	v, _ := r.UInt64()
	return v
}

func (r Reference) UInt64() (uint64, error) {
	if r.type_ == FBTUint {
		return r.data_.ReadUInt64(r.offset, r.parentWidth)
	} else {
		switch r.type_ {
		case FBTIndirectUInt:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			v, err := r.data_.ReadUInt64(ind, r.byteWidth)
			if err != nil {
				return 0, err
			}
			return v, nil
		case FBTInt:
			v, err := r.data_.ReadInt64(r.offset, r.parentWidth)
			if err != nil {
				return 0, err
			}
			return uint64(v), nil
		case FBTIndirectInt:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			v, err := r.data_.ReadInt64(ind, r.byteWidth)
			if err != nil {
				return 0, err
			}
			return uint64(v), nil
		case FBTFloat:
			v, err := r.data_.ReadDouble(r.offset, r.parentWidth)
			if err != nil {
				return 0, err
			}
			return uint64(v), nil
		case FBTIndirectFloat:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			v, err := r.data_.ReadDouble(ind, r.byteWidth)
			if err != nil {
				return 0, err
			}
			return uint64(v), nil
		case FBTNull:
			return 0, nil
		case FBTString:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			s, err := cstringBytesToString(r.data_[ind:])
			if err != nil {
				return 0, err
			}
			i, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return 0, err
			}
			return i, nil
		case FBTVector:
			vec, err := r.Vector()
			if err != nil {
				return 0, err
			}
			sz, err := vec.Size()
			if err != nil {
				return 0, err
			}
			return uint64(sz), nil
		case FBTBool:
			return r.data_.ReadUInt64(r.offset, r.parentWidth)
		default:
			return 0, ErrTypeDoesNotMatch
		}
	}
}
func (r Reference) AsFloat64() float64 {
	v, _ := r.Float64()
	return v
}

func (r Reference) Float64() (float64, error) {
	if r.type_ == FBTFloat {
		return r.data_.ReadDouble(r.offset, r.parentWidth)
	} else {
		switch r.type_ {
		case FBTIndirectFloat:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			return r.data_.ReadDouble(ind, r.byteWidth)
		case FBTUint:
			v, err := r.data_.ReadUInt64(r.offset, r.parentWidth)
			if err != nil {
				return 0, err
			}
			return float64(v), nil
		case FBTIndirectUInt:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			v, err := r.data_.ReadUInt64(ind, r.byteWidth)
			if err != nil {
				return 0, err
			}
			return float64(v), nil
		case FBTInt:
			v, err := r.data_.ReadInt64(r.offset, r.parentWidth)
			if err != nil {
				return 0, err
			}
			return float64(v), nil
		case FBTIndirectInt:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			v, err := r.data_.ReadInt64(ind, r.byteWidth)
			if err != nil {
				return 0, err
			}
			return float64(v), nil
		case FBTNull:
			return 0, nil
		case FBTString:
			ind, err := r.indirect()
			if err != nil {
				return 0, err
			}
			s, err := cstringBytesToString(r.data_[ind:])
			if err != nil {
				return 0, err
			}
			i, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0, err
			}
			return i, nil
		case FBTVector:
			vec, err := r.Vector()
			if err != nil {
				return 0, err
			}
			sz, err := vec.Size()
			if err != nil {
				return 0, err
			}
			return float64(sz), nil
		case FBTBool:
			v, err := r.data_.ReadUInt64(r.offset, r.parentWidth)
			if err != nil {
				return 0, err
			}
			return float64(v), nil
		default:
			return 0, ErrTypeDoesNotMatch
		}
	}
}

func (r Reference) AsFloat32() float32 {
	return float32(r.AsFloat64())
}

func (r Reference) asStringKey() (string, error) {
	if r.type_ == FBTKey {
		ind, err := r.indirect()
		if err != nil {
			return "", err
		}
		return unsafeReadCString(r.data_, ind)
	} else {
		return "", nil
	}
}
func (r Reference) AsKey() Key {
	v, err := r.Key()
	if err == ErrTypeDoesNotMatch {
		return EmptyKey()
	}
	return v
}

func (r Reference) Key() (Key, error) {
	if r.type_ != FBTKey {
		return Key{}, ErrTypeDoesNotMatch
	}
	ind, err := r.indirect()
	if err != nil {
		return Key{}, err
	}
	return Key{
		Object{
			buf:    r.data_,
			offset: ind,
		},
	}, nil
}
func (r Reference) AsStringRef() String {
	v, _ := r.StringRef()
	return v
}

func (r Reference) StringRef() (String, error) {
	if r.type_ != FBTString {
		return EmptyString(), ErrTypeDoesNotMatch
	}
	ind, err := r.indirect()
	if err != nil {
		return String{}, err
	}
	sz := Sized{
		Object{
			buf:       r.data_,
			offset:    ind,
			byteWidth: r.byteWidth,
		},
	}
	if r.hasExt {
		size, err := sz.Size()
		if err != nil {
			return EmptyString(), nil
		}
		sz.ext, _ = binary.Varint(r.data_[ind+size+1:]) //+1 for null byte
	}
	return String{sz}, nil
}
func (r Reference) AsBlob() Blob {
	v, _ := r.Blob()
	return v
}

func (r Reference) Blob() (Blob, error) {
	ind, err := r.indirect()
	if err != nil {
		return Blob{}, err
	}
	if r.type_ == FBTBlob || r.type_ == FBTString {
		sz := Sized{
			Object{
				buf:       r.data_,
				offset:    ind,
				byteWidth: r.byteWidth,
			},
		}
		if r.hasExt {
			size, err := sz.Size()
			if err != nil {
				return EmptyBlob(), nil
			}
			var n int
			sz.ext, n = binary.Varint(r.data_[ind+int(r.byteWidth)*size:])
			if n <= 0 {
				return EmptyBlob(), fmt.Errorf("failed to read ext")
			}
		}
		return Blob{sz}, nil
	} else {
		return EmptyBlob(), nil
	}
}

func (r Reference) AnyVector() (AnyVector, error) {
	if r.type_ == FBTVector {
		return r.Vector()
	} else if r.IsTypedVector() {
		return r.TypedVector()
	} else if r.IsFixedTypedVector() {
		return r.FixedTypedVector()
	} else {
		return nil, ErrTypeDoesNotMatch
	}
}

func (r Reference) AsAnyVector() AnyVector {
	v, _ := r.AnyVector()
	return v
}

func (r Reference) AsVector() Vector {
	v, _ := r.Vector()
	return v
}

func (r Reference) Vector() (Vector, error) {
	ind, err := r.indirect()
	if err != nil {
		return Vector{}, err
	}
	if r.type_ == FBTVector {
		sz := Sized{
			Object{
				buf:       r.data_,
				offset:    ind,
				byteWidth: r.byteWidth,
			},
		}
		if r.hasExt {
			size, err := sz.Size()
			if err != nil {
				return EmptyVector(), nil
			}
			// ind + body vector (byteWitdh * size) + type vector
			sz.ext, _ = binary.Varint(r.data_[ind+int(r.byteWidth)*size+size:])
		}
		return Vector{sz}, nil
	} else {
		return EmptyVector(), nil
	}
}
func (r Reference) AsTypedVector() TypedVector {
	v, _ := r.TypedVector()
	return v
}

func (r Reference) TypedVector() (TypedVector, error) {
	ind, err := r.indirect()
	if err != nil {
		return TypedVector{}, err
	}
	sz := Sized{
		Object{
			buf:       r.data_,
			offset:    ind,
			byteWidth: r.byteWidth,
		},
	}
	if r.hasExt {
		size, err := sz.Size()
		if err != nil {
			return EmptyTypedVector(), nil
		}
		sz.ext, _ = binary.Varint(r.data_[ind+int(r.byteWidth)*size:])
	}
	if r.IsTypedVector() {
		return TypedVector{
			Sized: sz,
			type_: ToTypedVectorElementType(r.type_),
		}, nil
	} else {
		return EmptyTypedVector(), nil
	}
}
func (r Reference) AsFixedTypedVector() FixedTypedVector {
	v, err := r.FixedTypedVector()
	if err == ErrTypeDoesNotMatch {
		return EmptyFixedTypedVector()
	}
	return v
}

func (r Reference) FixedTypedVector() (FixedTypedVector, error) {
	if !r.IsFixedTypedVector() {
		return FixedTypedVector{}, ErrTypeDoesNotMatch
	}
	ind, err := r.indirect()
	if err != nil {
		return FixedTypedVector{}, err
	}
	var l uint8
	vtype := ToFixedTypedVectorElementType(r.type_, &l)
	var ext int64
	if r.hasExt {
		ext, _ = binary.Varint(r.data_[ind+int(r.byteWidth)*int(l):])
	}
	return FixedTypedVector{
		Object: Object{
			buf:       r.data_,
			offset:    ind,
			byteWidth: r.byteWidth,
			ext:       ext,
		},
		type_: vtype,
		len_:  l,
	}, nil
}

func (r Reference) AsMap() Map {
	v, err := r.Map()
	if err == ErrTypeDoesNotMatch {
		return EmptyMap()
	}
	return v
}

func (r Reference) Map() (Map, error) {
	if r.type_ != FBTMap {
		return Map{}, ErrTypeDoesNotMatch
	}
	ind, err := r.indirect()
	if err != nil {
		return Map{}, err
	}
	sz := Sized{
		Object{
			buf:       r.data_,
			offset:    ind,
			byteWidth: r.byteWidth,
		},
	}
	if r.hasExt {
		size, err := sz.Size()
		if err != nil {
			return EmptyMap(), nil
		}

		numPrefixedData := 3
		keysOffset := ind - int(r.byteWidth)*numPrefixedData
		off, err := r.data_.Indirect(keysOffset, r.byteWidth)
		if err != nil {
			return EmptyMap(), fmt.Errorf("broken data: no ext for map")
		}
		bw, err := r.data_.ReadUInt64(keysOffset+int(r.byteWidth), r.byteWidth)
		if err != nil {
			return EmptyMap(), fmt.Errorf("broken data: no ext for map")
		}
		if bw <= 0 || bw > 8 {
			return EmptyMap(), ErrInvalidData
		}
		if off < 0 || len(r.data_) <= off {
			return EmptyMap(), ErrOutOfRange
		}
		sz.ext, _ = binary.Varint(r.data_[off+int(bw)*size:])
	}
	return Map{Vector{sz}}, nil
}

func (r Reference) MutateInt(i int64) error {
	switch r.type_ {
	case FBTInt:
		return r.data_.WriteInt64(r.offset, r.parentWidth, i)
	case FBTIndirectInt:
		ind, err := r.indirect()
		if err != nil {
			return err
		}
		return r.data_.WriteInt64(ind, r.byteWidth, i)
	case FBTUint:
		u := uint64(i)
		return r.data_.WriteUInt64(r.offset, r.parentWidth, u)
	case FBTIndirectUInt:
		ind, err := r.indirect()
		if err != nil {
			return err
		}
		u := uint64(i)
		return r.data_.WriteUInt64(ind, r.byteWidth, u)
	default:
		return ErrTypeDoesNotMatch
	}
}

func (r Reference) MutateUInt(u uint64) error {
	if r.type_ == FBTUint {
		return r.data_.WriteUInt64(r.offset, r.parentWidth, u)
	} else if r.type_ == FBTIndirectUInt {
		ind, err := r.indirect()
		if err != nil {
			return err
		}
		return r.data_.WriteUInt64(ind, r.byteWidth, u)
	} else if r.type_ == FBTInt {
		i := int64(u)
		return r.data_.WriteInt64(r.offset, r.parentWidth, i)
	} else if r.type_ == FBTIndirectInt {
		ind, err := r.indirect()
		if err != nil {
			return err
		}
		i := int64(u)
		return r.data_.WriteInt64(ind, r.byteWidth, i)
	} else {
		return ErrTypeDoesNotMatch
	}
}

func (r Reference) MutateFloat64(f float64) error {
	if r.type_ == FBTFloat {
		return r.data_.WriteFloat(r.offset, r.parentWidth, f)
	} else if r.type_ == FBTIndirectFloat {
		ind, err := r.indirect()
		if err != nil {
			return err
		}
		return r.data_.WriteFloat(ind, r.byteWidth, f)
	} else {
		return ErrTypeDoesNotMatch
	}
}

func (r Reference) MutateFloat32(f float32) error {
	if r.type_ == FBTFloat {
		return r.data_.WriteFloat(r.offset, r.parentWidth, float64(f))
	} else if r.type_ == FBTIndirectFloat {
		ind, err := r.indirect()
		if err != nil {
			return err
		}
		return r.data_.WriteFloat(ind, r.byteWidth, float64(f))
	} else {
		return ErrTypeDoesNotMatch
	}
}

func (r Reference) MutateString(s string) bool {
	if len(s) == 0 {
		return false
	}
	// This is very strict, could allow shorter strings, but that creates
	// garbage.
	// ... flexbuffers.h says so
	if len(s) != r.AsStringRef().SizeOrZero() {
		return false
	}
	data := *(*[]byte)(unsafe.Pointer(&s))
	if r.offset < 0 || len(r.data_) <= r.offset || len(r.data_) <= r.offset+len(data) {
		return false
	}
	copy(r.data_[r.offset:], data)
	r.data_[r.offset+len(data)] = 0 // NUL terminator
	return true
}
func (r Reference) Validate() (err error) {
	visited := make(map[int]struct{})
	return r.validate(visited)
}

func (r Reference) validate(visited map[int]struct{}) (err error) {
	if _, ok := visited[r.offset]; ok {
		return ErrRecursiveData
	}
	visited[r.offset] = struct{}{}

	_ = r.Ext()

	switch r.type_ {
	case FBTInt, FBTIndirectInt:
		_, err = r.Int64()
	case FBTUint, FBTIndirectUInt:
		_, err = r.UInt64()
	case FBTFloat, FBTIndirectFloat:
		_, err = r.Float64()
	case FBTKey:
		k, err := r.Key()
		if err != nil {
			return err
		}
		if strings.Index(k.StringValue(), "\x00") >= 0 {
			return fmt.Errorf("key contains null char")
		}
	case FBTString:
		s, err := r.StringRef()
		if err != nil {
			return err
		}
		if _, err := s.StringValue(); err != nil {
			return err
		}
	case FBTMap:
		m, err := r.Map()
		if err != nil {
			return err
		}
		keys, err := m.Keys()
		if err != nil {
			return err
		}
		var key Reference
		var value Reference
		sz, err := m.Size()
		if err != nil {
			return err
		}
		for i := 0; i < sz; i++ {
			if err := keys.AtRef(i, &key); err != nil {
				return err
			}
			if err := key.validate(visited); err != nil {
				return err
			}
			if err := m.AtRef(i, &value); err != nil {
				return err
			}
			if err := value.validate(visited); err != nil {
				return err
			}
		}
	case FBTVector:
		vec, err := r.Vector()
		if err != nil {
			return err
		}
		var v Reference
		sz, err := vec.Size()
		if err != nil {
			return err
		}
		for i := 0; i < sz; i++ {
			if err := vec.AtRef(i, &v); err != nil {
				return err
			}
			if v.offset == vec.offset {
				return ErrRecursiveData
			}
			if err := v.validate(visited); err != nil {
				return err
			}
		}
	default:
		anyVec, err := r.AnyVector()
		if err == nil {
			var v Reference
			sz, err := anyVec.Size()
			if err != nil {
				return err
			}
			for i := 0; i < sz; i++ {
				if err := anyVec.AtRef(i, &v); err != nil {
					return err
				}
				if err := v.validate(visited); err != nil {
					return err
				}
			}
			return nil
		}

		err = fmt.Errorf("type is invalid: %d", r.type_)
	}
	return
}

func (r Reference) Ext() int64 {
	if !r.hasExt {
		return 0
	}
	// TODO: optimize?
	switch {
	case r.IsString():
		return r.AsStringRef().Ext()
	case r.IsBlob():
		return r.AsBlob().Ext()
	case r.IsMap():
		return r.AsMap().Ext()
	case r.IsAnyVector():
		return r.AsAnyVector().Ext()
	default:
		return 0
	}
}
