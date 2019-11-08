package flexbuffers

import "C"
import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unsafe"
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
	parentWidth uint8
	byteWidth   uint8
	type_       Type
}

func (r Reference) String() string {
	var sb bytes.Buffer
	if err := r.WriteAsJson(&sb); err != nil {
		return err.Error()
	}
	return sb.String()
}

func (r Reference) writeJsonVector(l int, vec interface{ At(int) Reference }, w io.Writer) (err error) {
	if _, err = fmt.Fprintf(w, "["); err != nil {
		return
	}
	for i := 0; i < l; i++ {
		if err = vec.At(i).WriteAsJson(w); err != nil {
			return
		}
		if i != l-1 {
			if _, err = fmt.Fprintf(w, ","); err != nil {
				return
			}
		}
	}
	if _, err = fmt.Fprintf(w, "]"); err != nil {
		return
	}
	return
}

func (r Reference) WriteAsJson(w io.Writer) (err error) {
	switch r.type_ {
	case FBTNull:
		_, err = fmt.Fprintf(w, "null")
	case FBTInt, FBTIndirectInt:
		_, err = fmt.Fprintf(w, "%d", r.AsInt64())
	case FBTUint, FBTIndirectUInt:
		_, err = fmt.Fprintf(w, "%d", r.AsUInt64())
	case FBTFloat, FBTIndirectFloat:
		_, err = fmt.Fprintf(w, "%f", r.AsFloat64())
	case FBTKey:
		k, err := r.asStringKey()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\"%s\"", k)
	case FBTString:
		_, err = fmt.Fprintf(w, "\"%s\"", r.AsStringRef().StringValue())
	case FBTMap:
		m := r.AsMap()
		var keys TypedVector
		keys, err = m.Keys()
		if err != nil {
			return
		}
		values := m.Values()
		for i := 0; i < keys.Size(); i++ {
			k := keys.At(i)
			v := values.At(i)
			if err = k.WriteAsJson(w); err != nil {
				return
			}
			if _, err = fmt.Fprintf(w, ":"); err != nil {
				return
			}
			if err = v.WriteAsJson(w); err != nil {
				return
			}
		}
	case FBTVector:
		vec := r.AsVector()
		err = r.writeJsonVector(vec.Size(), vec, w)
	case FBTBlob:
		_, err = fmt.Fprintf(w, "\"%s\"", base64.StdEncoding.EncodeToString(r.AsBlob().Data()))
	case FBTBool:
		if r.AsBool() {
			_, err = fmt.Fprintf(w, "true")
		} else {
			_, err = fmt.Fprintf(w, "false")
		}
	default:
		if r.IsTypedVector() {
			vec := r.AsTypedVector()
			err = r.writeJsonVector(vec.Size(), vec, w)
		} else if r.IsFixedTypedVector() {
			vec := r.AsFixedTypedVector()
			err = r.writeJsonVector(int(vec.len_), vec, w)
		} else {
			return fmt.Errorf("unable to convert to json: type=%v", r.type_)
		}
	}
	return
}

func setReferenceFromPackedType(buf Raw, offset int, parentWidth uint8, packedType uint8, ref *Reference) {
	ref.data_ = buf
	ref.offset = offset
	ref.parentWidth = parentWidth
	ref.byteWidth = 1 << (packedType & 3)
	ref.type_ = Type(packedType >> 2)
}
func NewReferenceFromPackedType(buf Raw, offset int, parentWidth uint8, packedType uint8) Reference {
	return Reference{
		data_:       buf,
		offset:      offset,
		parentWidth: parentWidth,
		byteWidth:   1 << (packedType & 3),
		type_:       Type(packedType >> 2),
	}
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
			s := cstringBytesToString(r.data_[ind:])
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				panic("TODO")
			}
			return i, nil
		case FBTVector:
			return int64(r.AsVector().Size()), nil
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
			s := cstringBytesToString(r.data_[ind:])
			i, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return 0, err
			}
			return i, nil
		case FBTVector:
			v, err := r.Vector()
			if err != nil {
				return 0, err
			}
			return uint64(v.Size()), nil
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
			s := cstringBytesToString(r.data_[ind:])
			i, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0, err
			}
			return i, nil
		case FBTVector:
			v, err := r.Vector()
			if err != nil {
				return 0, err
			}
			return float64(v.Size()), nil
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
		return unsafeReadCString(r.data_, ind), nil
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
	return String{
		Sized{
			Object{
				buf:       r.data_,
				offset:    ind,
				byteWidth: r.byteWidth,
			},
		},
	}, nil
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
		return Blob{
			Sized{
				Object{
					buf:       r.data_,
					offset:    ind,
					byteWidth: r.byteWidth,
				},
			},
		}, nil
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

func (r Reference) AsVector() Vector {
	v, _ := r.Vector()
	return v
}

func (r Reference) Vector() (Vector, error) {
	ind, err := r.indirect()
	if err != nil {
		return Vector{}, err
	}
	if r.type_ == FBTVector || r.type_ == FBTMap {
		return Vector{
			Sized{
				Object{
					buf:       r.data_,
					offset:    ind,
					byteWidth: r.byteWidth,
				},
			},
		}, nil
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
	if r.IsTypedVector() {
		return TypedVector{
			Sized: Sized{
				Object{
					buf:       r.data_,
					offset:    ind,
					byteWidth: r.byteWidth,
				},
			},
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
	return FixedTypedVector{
		Object: Object{
			buf:       r.data_,
			offset:    ind,
			byteWidth: r.byteWidth,
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
	return Map{Vector{Sized{Object{
		buf:       r.data_,
		offset:    ind,
		byteWidth: r.byteWidth,
	}}}}, nil
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
	if len(s) != r.AsStringRef().Size() {
		return false
	}
	data := *(*[]byte)(unsafe.Pointer(&s))
	copy(r.data_[r.offset:], data)
	r.data_[r.offset+len(data)] = 0 // NUL terminator
	return true
}

func (r Reference) Validate() (err error) {
	switch r.type_ {
	case FBTInt, FBTIndirectInt:
		_, err = r.Int64()
	case FBTUint, FBTIndirectUInt:
		_, err = r.UInt64()
	case FBTFloat:
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
		s.StringValue()
	case FBTMap:
		m, err := r.Map()
		if err != nil {
			return err
		}
		keys, err := m.Keys()
		if err != nil {
			return err
		}
		var v Reference
		for i := 0; i < m.Size(); i++ {
			m.AtRef(i, &v)
			if err := v.Validate(); err != nil {
				return err
			}
			if err := keys.At(i).Validate(); err != nil {
				return err
			}
		}
	case FBTVector:
		vec, err := r.Vector()
		if err != nil {
			return err
		}
		var v Reference
		for i := 0; i < vec.Size(); i++ {
			vec.AtRef(i, &v)
			if err := v.Validate(); err != nil {
				return err
			}
		}
	default:
		anyVec, err := r.AnyVector()
		if err == nil {
			var v Reference
			for i := 0; i < anyVec.Size(); i++ {
				anyVec.AtRef(i, &v)
				if err := v.Validate(); err != nil {
					return err
				}
			}
			return nil
		}

		err = fmt.Errorf("type is invalid: %d", r.type_)
	}
	return
}
