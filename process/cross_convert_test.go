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

func (r *recordWriter) PushString(s string) error {
	r.history = append(r.history, recString(s))
	return nil
}

func (r *recordWriter) PushBlob(b []byte) error {
	panic("implement me")
}

func (r *recordWriter) PushInt(i int64) error {
	panic("implement me")
}

func (r *recordWriter) PushUint(u uint64) error {
	panic("implement me")
}

func (r *recordWriter) PushFloat(f float64) error {
	panic("implement me")
}

func (r *recordWriter) PushBool(b bool) error {
	panic("implement me")
}

func (r *recordWriter) PushNull() error {
	panic("implement me")
}

func (r *recordWriter) BeginArray() (int, error) {
	panic("implement me")
}

func (r *recordWriter) EndArray(int) error {
	panic("implement me")
}

func (r *recordWriter) BeginObject() (int, error) {
	panic("implement me")
}

func (r *recordWriter) EndObject(int) error {
	panic("implement me")
}

func (r *recordWriter) PushObjectKey(k string) error {
	panic("implement me")
}
