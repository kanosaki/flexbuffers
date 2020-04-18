package process

import (
	"testing"
)

func TestProductConvert(t *testing.T) {
	//readers := []struct {
	//	name string
	//}
	//writers := []struct {
	//	name string
	//}
	//cases := []func(w DocumentWriter){
	//	func(w DocumentWriter) {
	//
	//	},
	//}
	//
	//for _, reader := range readers {
	//	for _, writer := range writers {
	//		t.Run(fmt.Sprintf("%s-%s", reader.name, writer.name), func(t *testing.T) {
	//		})
	//	}
	//}
}

type recString string

type recordWriter struct {
	history []interface{}
}

func (r *recordWriter) PushString(ctx *Context, s string) error {
	r.history = append(r.history, recString(s))
	return nil
}

func (r *recordWriter) PushBlob(ctx *Context, b []byte) error {
	panic("implement me")
}

func (r *recordWriter) PushInt(ctx *Context, i int64) error {
	panic("implement me")
}

func (r *recordWriter) PushUint(ctx *Context, u uint64) error {
	panic("implement me")
}

func (r *recordWriter) PushFloat(ctx *Context, f float64) error {
	panic("implement me")
}

func (r *recordWriter) PushBool(ctx *Context, b bool) error {
	panic("implement me")
}

func (r *recordWriter) PushNull(*Context) error {
	panic("implement me")
}

func (r *recordWriter) BeginArray(*Context) (int, error) {
	panic("implement me")
}

func (r *recordWriter) EndArray(*Context, int) error {
	panic("implement me")
}

func (r *recordWriter) BeginObject(*Context) (int, error) {
	panic("implement me")
}

func (r *recordWriter) EndObject(*Context, int) error {
	panic("implement me")
}

func (r *recordWriter) PushObjectKey(ctx *Context, k string) error {
	panic("implement me")
}
