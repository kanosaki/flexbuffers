package process

import (
	"encoding"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"sync"

	"flexbuffers/pkg/unsafeutil"
)

type SelfMarshaller interface {
	Output(documentWriter DocumentWriter) error
}

type ObjectReader struct {
	Output DocumentWriter
}

func (r *ObjectReader) Read(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()
	return r.reflectValue(reflect.ValueOf(v))
}

func (r *ObjectReader) reflectValue(v reflect.Value) error {
	fn := valueEncoder(v)
	return fn(r, v)
}

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(v.Type())
}

func invalidValueEncoder(r *ObjectReader, v reflect.Value) error {
	return r.Output.PushNull(nil)
}

var (
	selfMarshallerType = reflect.TypeOf((*SelfMarshaller)(nil)).Elem()
	textMarshalerType  = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	encoderCache       sync.Map // map[reflect.Type]encoderFunc
)

type encodeState struct {
	o DocumentWriter
}

type encOpts struct {
}

type encoderFunc func(e *ObjectReader, v reflect.Value) error

func typeEncoder(t reflect.Type) encoderFunc {
	if fi, ok := encoderCache.Load(t); ok {
		return fi.(encoderFunc)
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	var (
		wg sync.WaitGroup
		f  encoderFunc
	)
	wg.Add(1)
	fi, loaded := encoderCache.LoadOrStore(t, encoderFunc(func(r *ObjectReader, v reflect.Value) error {
		wg.Wait()
		return f(r, v)
	}))
	if loaded {
		return fi.(encoderFunc)
	}

	// Compute the real encoder and replace the indirect func with it.
	f = newTypeEncoder(t, true)
	wg.Done()
	encoderCache.Store(t, f)
	return f
}

// newTypeEncoder constructs an encoderFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeEncoder(t reflect.Type, allowAddr bool) encoderFunc {
	if t.Implements(selfMarshallerType) {
		return marshalerEncoder
	}
	if t.Kind() != reflect.Ptr && allowAddr && reflect.PtrTo(t).Implements(selfMarshallerType) {
		return newCondAddrEncoder(addrMarshalerEncoder, newTypeEncoder(t, false))
	}

	if t.Implements(textMarshalerType) {
		return textMarshalerEncoder
	}
	if t.Kind() != reflect.Ptr && allowAddr && reflect.PtrTo(t).Implements(textMarshalerType) {
		return newCondAddrEncoder(addrTextMarshalerEncoder, newTypeEncoder(t, false))
	}

	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32:
		return float32Encoder
	case reflect.Float64:
		return float64Encoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct:
		return newStructEncoder(t)
	case reflect.Map:
		return newMapEncoder(t)
	case reflect.Slice:
		return newSliceEncoder(t)
	case reflect.Array:
		return newArrayEncoder(t)
	case reflect.Ptr:
		return newPtrEncoder(t)
	default:
		return unsupportedTypeEncoder
	}
}

func marshalerEncoder(r *ObjectReader, v reflect.Value) error {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return r.Output.PushNull(nil)
	}
	m, ok := v.Interface().(SelfMarshaller)
	if !ok {
		return r.Output.PushNull(nil)
	}
	return m.Output(r.Output)
}

func addrMarshalerEncoder(r *ObjectReader, v reflect.Value) error {
	va := v.Addr()
	if va.IsNil() {
		return r.Output.PushNull(nil)
	}
	m, ok := va.Interface().(SelfMarshaller)
	if !ok {
		return r.Output.PushNull(nil)
	}
	return m.Output(r.Output)
}

func stringEncoder(r *ObjectReader, v reflect.Value) error {
	return r.Output.PushString(nil, v.String())
}

func interfaceEncoder(r *ObjectReader, v reflect.Value) error {
	if v.IsNil() {
		return r.Output.PushNull(nil)
	}
	return r.reflectValue(v.Elem())
}

func unsupportedTypeEncoder(r *ObjectReader, v reflect.Value) error {
	return fmt.Errorf("unsupported type: %v", v.Type())
}

type structEncoder struct {
	fields structFields
}

func newStructEncoder(t reflect.Type) encoderFunc {
	se := structEncoder{fields: cachedTypeFields(t)}
	return se.encode
}

type reflectWithString struct {
	v reflect.Value
	s string
}

func (w *reflectWithString) resolve() error {
	if w.v.Kind() == reflect.String {
		w.s = w.v.String()
		return nil
	}
	if tm, ok := w.v.Interface().(encoding.TextMarshaler); ok {
		buf, err := tm.MarshalText()
		w.s = string(buf)
		return err
	}
	switch w.v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.s = strconv.FormatInt(w.v.Int(), 10)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		w.s = strconv.FormatUint(w.v.Uint(), 10)
		return nil
	}
	panic("unexpected map key type")
}

type mapEncoder struct {
	elemEnc encoderFunc
}

func (me mapEncoder) encode(r *ObjectReader, v reflect.Value) error {
	if v.IsNil() {
		return r.Output.PushNull(nil)
	}

	// Extract and sort the keys.
	keys := v.MapKeys()
	sv := make([]reflectWithString, len(keys))
	for i, v := range keys {
		sv[i].v = v
		if err := sv[i].resolve(); err != nil {
			return err
		}
	}

	ptr, err := r.Output.BeginObject(nil)
	if err != nil {
		return err
	}
	for _, kv := range sv {
		if err := r.Output.PushObjectKey(nil, kv.s); err != nil {
			return err
		}

		if err := me.elemEnc(r, v.MapIndex(kv.v)); err != nil {
			return err
		}
	}
	return r.Output.EndObject(nil, ptr)
}

func newMapEncoder(t reflect.Type) encoderFunc {
	switch t.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		if !t.Key().Implements(textMarshalerType) {
			return unsupportedTypeEncoder
		}
	}
	me := mapEncoder{typeEncoder(t.Elem())}
	return me.encode
}

func encodeByteSlice(r *ObjectReader, v reflect.Value) error {
	if v.IsNil() {
		return r.Output.PushNull(nil)
	}
	s := v.Bytes()
	return r.Output.PushBlob(nil, s)
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se sliceEncoder) encode(r *ObjectReader, v reflect.Value) error {
	if v.IsNil() {
		return r.Output.PushNull(nil)
	}
	return se.arrayEnc(r, v)
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		p := reflect.PtrTo(t.Elem())
		if !p.Implements(selfMarshallerType) {
			return encodeByteSlice
		}
	}
	enc := sliceEncoder{newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae arrayEncoder) encode(r *ObjectReader, v reflect.Value) error {
	ptr, err := r.Output.BeginArray(nil)
	if err != nil {
		return err
	}
	n := v.Len()
	for i := 0; i < n; i++ {
		if err := ae.elemEnc(r, v.Index(i)); err != nil {
			return err
		}
	}
	return r.Output.EndArray(nil, ptr)
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := arrayEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe ptrEncoder) encode(r *ObjectReader, v reflect.Value) error {
	if v.IsNil() {
		return r.Output.PushNull(nil)
	}
	return pe.elemEnc(r, v.Elem())
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := ptrEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type condAddrEncoder struct {
	canAddrEnc, elseEnc encoderFunc
}

func (ce condAddrEncoder) encode(r *ObjectReader, v reflect.Value) error {
	if v.CanAddr() {
		return ce.canAddrEnc(r, v)
	} else {
		return ce.elseEnc(r, v)
	}
}

// newCondAddrEncoder returns an encoder that checks whether its value
// CanAddr and delegates to canAddrEnc if so, else to elseEnc.
func newCondAddrEncoder(canAddrEnc, elseEnc encoderFunc) encoderFunc {
	enc := condAddrEncoder{canAddrEnc: canAddrEnc, elseEnc: elseEnc}
	return enc.encode
}

func boolEncoder(r *ObjectReader, v reflect.Value) error {
	return r.Output.PushBool(nil, v.Bool())
}

func intEncoder(r *ObjectReader, v reflect.Value) error {
	return r.Output.PushInt(nil, v.Int())
}

func uintEncoder(r *ObjectReader, v reflect.Value) error {
	u := v.Uint()
	if u > math.MaxInt64 {
		return fmt.Errorf("cannot encode %d: out of boundary", u)
	}
	return r.Output.PushInt(nil, int64(u))
}

type floatEncoder int // number of bits

func (bits floatEncoder) encode(r *ObjectReader, v reflect.Value) error {
	return r.Output.PushFloat(nil, v.Float())
}

var (
	float32Encoder = (floatEncoder(32)).encode
	float64Encoder = (floatEncoder(64)).encode
)

func textMarshalerEncoder(r *ObjectReader, v reflect.Value) error {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return r.Output.PushNull(nil)
	}
	m := v.Interface().(encoding.TextMarshaler)
	b, err := m.MarshalText()
	if err != nil {
		return err
	}
	// TODO: or key?
	return r.Output.PushString(nil, unsafeutil.B2S(b))
}

func addrTextMarshalerEncoder(r *ObjectReader, v reflect.Value) error {
	va := v.Addr()
	if va.IsNil() {
		return r.Output.PushNull(nil)
	}
	m := va.Interface().(encoding.TextMarshaler)
	b, err := m.MarshalText()
	if err != nil {
		return err
	}
	return r.Output.PushString(nil, unsafeutil.B2S(b))
}
