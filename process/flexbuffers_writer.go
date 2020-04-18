package process

import (
	"flexbuffers"
	"flexbuffers/pkg/unsafeutil"
)

type FlexbuffersWriter struct {
	b *flexbuffers.Builder
}

func (w *FlexbuffersWriter) PushString(ctx *Context, s string) error {
	_ = w.b.StringValue(s)
	return nil
}

func (w *FlexbuffersWriter) PushBlob(ctx *Context, b []byte) error {
	_ = w.b.Blob(b)
	return nil
}

func (w *FlexbuffersWriter) PushInt(ctx *Context, i int64) error {
	w.b.Int(i)
	return nil
}

func (w *FlexbuffersWriter) PushUint(ctx *Context, u uint64) error {
	w.b.UInt(u)
	return nil
}

func (w *FlexbuffersWriter) PushFloat(ctx *Context, f float64) error {
	w.b.Float64(f)
	return nil
}

func (w *FlexbuffersWriter) PushBool(ctx *Context, b bool) error {
	w.b.Bool(b)
	return nil
}

func (w *FlexbuffersWriter) PushNull(*Context) error {
	w.b.Null()
	return nil
}

func (w *FlexbuffersWriter) BeginArray(*Context) (int, error) {
	return w.b.StartVector(), nil
}

func (w *FlexbuffersWriter) EndArray(ctx *Context, id int) error {
	// TODO: make typed or fixed automatically
	_, err := w.b.EndVector(id, false, false)
	return err
}

func (w *FlexbuffersWriter) BeginObject(*Context) (int, error) {
	return w.b.StartMap(), nil
}

func (w *FlexbuffersWriter) EndObject(ctx *Context, id int) error {
	_, err := w.b.EndMap(id)
	return err
}

func (w *FlexbuffersWriter) PushObjectKey(ctx *Context, k string) error {
	w.b.Key(unsafeutil.S2B(k))
	return nil
}
