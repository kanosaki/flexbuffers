package flexbuffers

import "C"
import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"unsafe"
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
		_, err = fmt.Fprintf(w, "\"%s\"", r.asStringKey())
	case FBTString:
		_, err = fmt.Fprintf(w, "\"%s\"", r.AsString().StringValue())
	case FBTMap:
		m := r.AsMap()
		keys := m.Keys()
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
	if r.type_ == FBTBool {
		return r.data_.ReadUInt64(r.offset, r.parentWidth) != 0
	} else {
		return r.AsUInt64() != 0
	}
}

func (r Reference) indirect() int {
	return r.data_.Indirect(r.offset, r.parentWidth)
}

func (r Reference) AsInt64() int64 {
	if r.type_ == FBTInt {
		return r.data_.ReadInt64(r.offset, r.parentWidth)
	} else {
		switch r.type_ {
		case FBTIndirectInt:
			return r.data_.ReadInt64(r.indirect(), r.byteWidth)
		case FBTUint:
			return int64(r.data_.ReadUInt64(r.offset, r.parentWidth))
		case FBTIndirectUInt:
			return int64(r.data_.ReadUInt64(r.indirect(), r.byteWidth))
		case FBTFloat:
			return int64(r.data_.ReadDouble(r.offset, r.parentWidth))
		case FBTIndirectFloat:
			return int64(r.data_.ReadDouble(r.indirect(), r.byteWidth))
		case FBTNull:
			return 0
		case FBTString:
			s := cstringBytesToString(r.data_[r.indirect():])
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				panic("TODO")
			}
			return i
		case FBTVector:
			return int64(r.AsVector().Size())
		case FBTBool:
			return r.data_.ReadInt64(r.offset, r.parentWidth)
		default:
			return 0
		}
	}
}

func (r Reference) AsUInt64() uint64 {
	if r.type_ == FBTUint {
		return r.data_.ReadUInt64(r.offset, r.parentWidth)
	} else {
		switch r.type_ {
		case FBTIndirectUInt:
			return uint64(r.data_.ReadUInt64(r.indirect(), r.byteWidth))
		case FBTInt:
			return uint64(r.data_.ReadInt64(r.offset, r.parentWidth))
		case FBTIndirectInt:
			return uint64(r.data_.ReadInt64(r.indirect(), r.byteWidth))
		case FBTFloat:
			return uint64(r.data_.ReadDouble(r.offset, r.parentWidth))
		case FBTIndirectFloat:
			return uint64(r.data_.ReadDouble(r.indirect(), r.byteWidth))
		case FBTNull:
			return 0
		case FBTString:
			s := cstringBytesToString(r.data_[r.indirect():])
			i, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				panic("TODO")
			}
			return i
		case FBTVector:
			return uint64(r.AsVector().Size())
		case FBTBool:
			return r.data_.ReadUInt64(r.offset, r.parentWidth)
		default:
			return 0
		}
	}
}

func (r Reference) AsFloat64() float64 {
	if r.type_ == FBTFloat {
		return r.data_.ReadDouble(r.offset, r.parentWidth)
	} else {
		switch r.type_ {
		case FBTIndirectFloat:
			return r.data_.ReadDouble(r.indirect(), r.byteWidth)
		case FBTUint:
			return float64(r.data_.ReadUInt64(r.offset, r.parentWidth))
		case FBTIndirectUInt:
			return float64(r.data_.ReadUInt64(r.indirect(), r.byteWidth))
		case FBTInt:
			return float64(r.data_.ReadInt64(r.offset, r.parentWidth))
		case FBTIndirectInt:
			return float64(r.data_.ReadInt64(r.indirect(), r.byteWidth))
		case FBTNull:
			return 0
		case FBTString:
			s := cstringBytesToString(r.data_[r.indirect():])
			i, err := strconv.ParseFloat(s, 64)
			if err != nil {
				panic("TODO")
			}
			return i
		case FBTVector:
			return float64(r.AsVector().Size())
		case FBTBool:
			return float64(r.data_.ReadUInt64(r.offset, r.parentWidth))
		default:
			return 0
		}
	}
}

func (r Reference) AsFloat32() float32 {
	return float32(r.AsFloat64())
}

func (r Reference) asStringKey() string {
	if r.type_ == FBTKey {
		ind := r.indirect()
		return unsafeReadCString(r.data_, ind)
	} else {
		return ""
	}
}

func (r Reference) AsKey() Key {
	if r.type_ == FBTKey {
		ind := r.indirect()
		return Key{
			Object{
				buf:    r.data_,
				offset: ind,
			},
		}
	} else {
		return EmptyKey()
	}
}

func (r Reference) AsString() String {
	if r.type_ == FBTString {
		ind := r.indirect()
		return String{
			Sized{
				Object{
					buf:       r.data_,
					offset:    ind,
					byteWidth: r.byteWidth,
				},
			},
		}
	} else {
		return EmptyString()
	}
}

func (r Reference) AsBlob() Blob {
	if r.type_ == FBTBlob || r.type_ == FBTString {
		return Blob{
			Sized{
				Object{
					buf:       r.data_,
					offset:    r.indirect(),
					byteWidth: r.byteWidth,
				},
			},
		}
	} else {
		return EmptyBlob()
	}
}

func (r Reference) AsVector() Vector {
	if r.type_ == FBTVector || r.type_ == FBTMap {
		return Vector{
			Sized{
				Object{
					buf:       r.data_,
					offset:    r.indirect(),
					byteWidth: r.byteWidth,
				},
			},
		}
	} else {
		return EmptyVector()
	}
}

func (r Reference) AsTypedVector() TypedVector {
	if r.IsTypedVector() {
		return TypedVector{
			Sized: Sized{
				Object{
					buf:       r.data_,
					offset:    r.indirect(),
					byteWidth: r.byteWidth,
				},
			},
			type_: ToTypedVectorElementType(r.type_),
		}
	} else {
		return EmptyTypedVector()
	}
}

func (r Reference) AsFixedTypedVector() FixedTypedVector {
	if r.IsFixedTypedVector() {
		var l uint8
		vtype := ToFixedTypedVectorElementType(r.type_, &l)
		return FixedTypedVector{
			Object: Object{
				buf:       r.data_,
				offset:    r.indirect(),
				byteWidth: r.byteWidth,
			},
			type_: vtype,
			len_:  l,
		}
	} else {
		return EmptyFixedTypedVector()
	}
}

func (r Reference) AsMap() Map {
	if r.type_ == FBTMap {
		return Map{Vector{Sized{Object{
			buf:       r.data_,
			offset:    r.indirect(),
			byteWidth: r.byteWidth,
		}}}}
	} else {
		return EmptyMap()
	}
}

func (r Reference) MutateInt(i int64) bool {
	if r.type_ == FBTInt {
		return r.data_.WriteInt64(r.offset, r.parentWidth, i)
	} else if r.type_ == FBTIndirectInt {
		return r.data_.WriteInt64(r.indirect(), r.byteWidth, i)
	} else if r.type_ == FBTUint {
		u := uint64(i)
		return r.data_.WriteUInt64(r.offset, r.parentWidth, u)
	} else if r.type_ == FBTIndirectUInt {
		u := uint64(i)
		return r.data_.WriteUInt64(r.indirect(), r.byteWidth, u)
	} else {
		return false
	}
}

func (r Reference) MutateUInt(u uint64) bool {
	if r.type_ == FBTUint {
		return r.data_.WriteUInt64(r.offset, r.parentWidth, u)
	} else if r.type_ == FBTIndirectUInt {
		return r.data_.WriteUInt64(r.indirect(), r.byteWidth, u)
	} else if r.type_ == FBTInt {
		i := int64(u)
		return r.data_.WriteInt64(r.offset, r.parentWidth, i)
	} else if r.type_ == FBTIndirectInt {
		i := int64(u)
		return r.data_.WriteInt64(r.indirect(), r.byteWidth, i)
	} else {
		return false
	}
}

func (r Reference) MutateFloat64(f float64) bool {
	if r.type_ == FBTFloat {
		return r.data_.WriteFloat(r.offset, r.parentWidth, f)
	} else if r.type_ == FBTIndirectFloat {
		return r.data_.WriteFloat(r.indirect(), r.byteWidth, f)
	} else {
		return false
	}
}

func (r Reference) MutateFloat32(f float32) bool {
	if r.type_ == FBTFloat {
		return r.data_.WriteFloat(r.offset, r.parentWidth, float64(f))
	} else if r.type_ == FBTIndirectFloat {
		return r.data_.WriteFloat(r.indirect(), r.byteWidth, float64(f))
	} else {
		return false
	}
}

func (r Reference) MutateString(s string) bool {
	if len(s) == 0 {
		return false
	}
	// This is very strict, could allow shorter strings, but that creates
	// garbage.
	// ... flexbuffers.h says so
	if len(s) != r.AsString().Size() {
		return false
	}
	data := *(*[]byte)(unsafe.Pointer(&s))
	copy(r.data_[r.offset:], data)
	r.data_[r.offset+len(data)] = 0 // NUL terminator
	return true
}
