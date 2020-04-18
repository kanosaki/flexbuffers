package process

import (
	"bytes"
	"encoding"
	"fmt"
	"reflect"

	"flexbuffers/pkg/unsafeutil"
)

type ObjectWriter struct {
	target interface{}

	stack       []valueSetter
	selectedKey *string
	// options
	DisallowUnknownField bool
}

type Unmarshaler interface {
	DocumentWriter
}

func NewObjectWriter(target interface{}) (*ObjectWriter, error) {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, fmt.Errorf("cannot unmarshal into non-pointer target")
	}
	return &ObjectWriter{
		target: target,
		stack: []valueSetter{{
			target: rv,
		}},
		DisallowUnknownField: true,
	}, nil
}

func (o *ObjectWriter) PushString(ctx *Context, s string) error {
	var setter valueSetter
	um, err := o.prepareTarget(&setter, false)
	if err != nil {
		return err
	}
	if um != nil {
		return um.PushString(ctx, s)
	}
	return setter.SetString(s)
}

func (o *ObjectWriter) PushBlob(ctx *Context, b []byte) error {
	panic("implement me")
}

func (o *ObjectWriter) PushInt(ctx *Context, i int64) error {
	var setter valueSetter
	um, err := o.prepareTarget(&setter, false)
	if err != nil {
		return err
	}
	if um != nil {
		return um.PushInt(ctx, i)
	}
	return setter.SetInt(i)
}

func (o *ObjectWriter) PushUint(ctx *Context, u uint64) error {
	panic("implement me")
}

func (o *ObjectWriter) PushFloat(ctx *Context, f float64) error {
	panic("implement me")
}

func (o *ObjectWriter) PushBool(ctx *Context, b bool) error {
	panic("implement me")
}

func (o *ObjectWriter) PushNull(*Context) error {
	panic("implement me")
}

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func (o *ObjectWriter) BeginArray(*Context) (int, error) {
	var setter valueSetter
	um, err := o.prepareTarget(&setter, false)
	if err != nil {
		return 0, err
	}
	if um != nil {
		return um.BeginArray(nil)
	}
	o.stack = append(o.stack, setter)
	o.selectedKey = nil
	return 0, setter.BeginArray()
}

func (o *ObjectWriter) EndArray(*Context, int) error {
	popd := o.stack[len(o.stack)-1]
	o.stack = o.stack[:len(o.stack)-1]
	return popd.EndArray()
}

func (o *ObjectWriter) BeginObject(*Context) (int, error) {
	var setter valueSetter
	um, err := o.prepareTarget(&setter, false)
	if err != nil {
		return 0, err
	}
	if um != nil {
		return um.BeginObject(nil)
	}
	if err := setter.BeginObject(); err != nil {
		return 0, err
	}
	o.stack = append(o.stack, setter)
	o.selectedKey = nil
	return len(o.stack), nil
}

func (o *ObjectWriter) EndObject(*Context, int) error {
	o.selectedKey = nil
	popd := o.stack[len(o.stack)-1]
	o.stack = o.stack[:len(o.stack)-1]
	return popd.EndObject()
}

func (o *ObjectWriter) prepareTarget(setter *valueSetter, decodingNil bool) (Unmarshaler, error) {
	vs := o.stack[len(o.stack)-1]
	v := vs.target
	if o.selectedKey == nil {
		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			*setter = valueSetter{
				parent: v,
			}
		default:
			*setter = valueSetter{
				target: v,
			}
		}
		return nil, nil
	}
	key := *o.selectedKey

	if !v.IsValid() {
		panic("v invalid")
	}
	var unmarshaller Unmarshaler
	unmarshaller, v = indirect(v, decodingNil)
	if unmarshaller != nil {
		return unmarshaller, nil
	}

	var fields structFields
	t := v.Type()
	switch v.Kind() {
	case reflect.Map:
		switch t.Key().Kind() {
		case reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		default:
			if !reflect.PtrTo(t.Key()).Implements(textUnmarshalerType) {
				return nil, fmt.Errorf("key type is invalid")
			}
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		}
	case reflect.Struct:
		fields = cachedTypeFields(t)
	default:
		return nil, fmt.Errorf("cannot set to %+v(kind: %v)", v, v.Kind())
	}
	// Figure out field corresponding to key.
	var subv reflect.Value

	var mapElem reflect.Value

	if v.Kind() == reflect.Map {
		elemType := t.Elem()
		if !mapElem.IsValid() {
			mapElem = reflect.New(elemType).Elem()
		} else {
			mapElem.Set(reflect.Zero(elemType))
		}
		subv = mapElem
	} else {
		var f *field
		if i, ok := fields.nameIndex[key]; ok {
			f = &fields.list[i]
		} else {
			for i := range fields.list {
				ff := &fields.list[i]
				if bytes.EqualFold(unsafeutil.S2B(ff.name), unsafeutil.S2B(key)) {
					f = ff
					break
				}
			}
		}
		if f != nil {
			subv = v
			for _, i := range f.index {
				if subv.Kind() == reflect.Ptr {
					if subv.IsNil() {
						// If a struct embeds a pointer to an unexported type,
						// it is not possible to set a newly allocated value
						// since the field is unexported.
						//
						// See https://golang.org/issue/21357
						if !subv.CanSet() {
							return nil, fmt.Errorf("cannot set embedded pointer to unexported struct: %v", subv.Type().Elem())
							// Invalidate subv to ensure d.value(subv) skips over
							// the JSON value without assigning it to subv.
						}
						subv.Set(reflect.New(subv.Type().Elem()))
					}
					subv = subv.Elem()
				}
				subv = subv.Field(i)
			}
		} else if o.DisallowUnknownField {
			return nil, fmt.Errorf("unknwon field %q", key)
		}
	}
	*setter = valueSetter{
		target:      subv,
		parent:      v,
		selectedKey: o.selectedKey,
	}
	return nil, nil
}

func (o *ObjectWriter) PushObjectKey(ctx *Context, k string) error {
	o.selectedKey = &k
	return nil

}

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// if it encounters an Unmarshaler, indirect stops and returns that.
// if decodingNull is true, indirect stops at the last pointer so it can be set to nil.
func indirect(v reflect.Value, decodingNull bool) (Unmarshaler, reflect.Value) {
	// Issue #24153 indicates that it is generally not a guaranteed property
	// that you may round-trip a reflect.Value by calling Value.Addr().Elem()
	// and expect the value to still be settable for values derived from
	// unexported embedded struct fields.
	//
	// The logic below effectively does this when it first addresses the value
	// (to satisfy possible pointer methods) and continues to dereference
	// subsequent pointers as necessary.
	//
	// After the first round-trip, we set v back to the original value to
	// preserve the original RW flags contained in reflect.Value.
	v0 := v
	haveAddr := false

	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		haveAddr = true
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				haveAddr = false
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
			break
		}

		// Prevent infinite loop if v is an interface pointing to its own address:
		//     var v interface{}
		//     v = &v
		if v.Elem().Kind() == reflect.Interface && v.Elem().Elem() == v {
			v = v.Elem()
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if v.Type().NumMethod() > 0 && v.CanInterface() {
			if u, ok := v.Interface().(Unmarshaler); ok {
				return u, reflect.Value{}
			}
		}

		if haveAddr {
			v = v0 // restore original value after round-trip Value.Addr().Elem()
			haveAddr = false
		} else {
			v = v.Elem()
		}
	}
	return nil, v
}

type valueSetter struct {
	// if true, target is a pointer to original value
	ptrMode     bool
	target      reflect.Value
	parent      reflect.Value
	grandparent reflect.Value
	selectedKey *string
}

func applyValueToPtrs(ptrs, val reflect.Value, convert bool) reflect.Value {
	elemType := ptrs.Elem().Kind()
	if elemType == reflect.Ptr {
		ret := reflect.New(ptrs.Type().Elem())
		ret.Set(applyValueToPtrs(ptrs.Elem(), val, convert))
		return ret
	} else {
		v := val
		if convert {
			v = val.Convert(ptrs.Elem().Type())
		}
		newPtr := reflect.New(ptrs.Elem().Type())
		newPtr.Set(v)
		ptrs.Set(newPtr)
		return ptrs
	}
}

func (vs *valueSetter) SetInt(i int64) error {
	v := vs.target
	if !v.IsValid() {
		et := vs.parent.Type().Elem()
		vs.parent.Set(reflect.Append(vs.parent, reflect.ValueOf(i).Convert(et)))
		return nil
	}

	var unmarshaller Unmarshaler
	unmarshaller, v = indirect(v, false)
	if unmarshaller != nil {
		return unmarshaller.PushInt(nil, i)
	}

	switch v.Kind() {
	case reflect.Interface:
		if v.NumMethod() != 0 {
			return fmt.Errorf("cannot set int value into %+v", v)
		}
		vs.set(v, reflect.ValueOf(i))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.OverflowInt(i) {
			return fmt.Errorf("%d overflows %v", i, v)
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if i < 0 {
			return fmt.Errorf("cannot set negative number (%d) to uint field (%v)", i, v)
		}
		if v.OverflowUint(uint64(i)) {
			return fmt.Errorf("%d overflows %v", i, v)
		}
		v.SetUint(uint64(i))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(float64(i))
	case reflect.Slice, reflect.Array:
		et := v.Type().Elem()
		v.Set(reflect.Append(v, reflect.ValueOf(i).Convert(et)))
	default:
		return fmt.Errorf("canno set int value into %+v", v)
	}
	return nil
}

func (vs *valueSetter) SetString(s string) error {
	v := vs.target

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Interface:
		if v.NumMethod() != 0 {
			return fmt.Errorf("cannot set string value into %+v", v)
		}
		vs.set(v, reflect.ValueOf(s))
	case reflect.Slice, reflect.Array:
		v.Set(reflect.Append(v, reflect.ValueOf(s)))
	default:
		return fmt.Errorf("canno set string value into %+v", v)
	}
	return nil
}

func (vs *valueSetter) BeginArray() error {
	return nil
}

func (vs *valueSetter) EndArray() error {
	return nil
}

func (vs *valueSetter) BeginObject() error {
	v := vs.target
	if !v.IsValid() { // append mode
		p := vs.parent
		et := p.Type().Elem()
		var t reflect.Value
		switch et.Kind() {
		case reflect.Map:
			t = reflect.MakeMap(et)
		case reflect.Struct:
			t = reflect.New(et)
			vs.ptrMode = true
		case reflect.Ptr:
			t = reflect.New(et.Elem())
		case reflect.Slice:
			t = reflect.MakeSlice(et, 0, 8)
		}
		vs.target = t
	} else {
		switch v.Kind() {
		case reflect.Map:
			if v.IsNil() {
				vs.set(v, reflect.MakeMap(v.Type()))
			}
		}
	}
	return nil
}

func (vs *valueSetter) EndObject() error {
	switch vs.parent.Kind() {
	case reflect.Slice, reflect.Array:
		if vs.ptrMode {
			vs.parent.Set(reflect.Append(vs.parent, vs.target.Elem()))
		} else {
			vs.parent.Set(reflect.Append(vs.parent, vs.target))
		}
	}
	return nil
}

func (vs *valueSetter) set(target, val reflect.Value) {
	outer := vs.parent
	if outer.Kind() == reflect.Map {
		outer.SetMapIndex(reflect.ValueOf(*vs.selectedKey), val)
	} else {
		target.Set(val)
	}
}
