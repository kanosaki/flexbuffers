package process

import (
	"flexbuffers"
	"flexbuffers/pkg/unsafeutil"
)

type FlexbuffersWriter struct {
	b *flexbuffers.Builder
}

func (w *FlexbuffersWriter) PushString(s string) error {
	_ = w.b.StringValue(s)
	return nil
}

func (w *FlexbuffersWriter) PushBlob(blb []byte) error {
	_ = w.b.Blob(blb)
	return nil
}

func (w *FlexbuffersWriter) PushInt(i int64) error {
	w.b.Int(i)
	return nil
}

func (w *FlexbuffersWriter) PushUint(u uint64) error {
	w.b.UInt(u)
	return nil
}

func (w *FlexbuffersWriter) PushFloat(f float64) error {
	w.b.Float64(f)
	return nil
}

func (w *FlexbuffersWriter) PushBool(tf bool) error {
	w.b.Bool(tf)
	return nil
}

func (w *FlexbuffersWriter) PushNull() error {
	w.b.Null()
	return nil
}

func (w *FlexbuffersWriter) BeginArray() (int, error) {
	return w.b.StartVector(), nil
}

func (w *FlexbuffersWriter) EndArray(ptr int) error {
	// TODO: make typed or fixed automatically
	_, err := w.b.EndVector(ptr, false, false)
	return err
}

func (w *FlexbuffersWriter) BeginObject() (int, error) {
	return w.b.StartMap(), nil
}

func (w *FlexbuffersWriter) EndObject(ptr int) error {
	_, err := w.b.EndMap(ptr)
	return err
}

func (w *FlexbuffersWriter) PushObjectKey(k string) error {
	w.b.Key(unsafeutil.S2B(k))
	return nil
}
