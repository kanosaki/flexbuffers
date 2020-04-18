package process

import (
	"encoding/base64"
	"fmt"
	"io"

	"flexbuffers"
)

type JsonWriter struct {
	Output     io.Writer
	elemIndex  []int
	keyWritten bool
}

func (j *JsonWriter) PushString(ctx *Context, s string) (err error) {
	if err := j.preElem(); err != nil {
		return err
	}
	buf := make([]byte, 0, len(s))
	buf = flexbuffers.EscapeJSONString(buf, s)
	_, err = j.Output.Write(buf)
	if err != nil {
		return
	}
	j.incrementElemIndex()
	return
}

func (j *JsonWriter) PushBlob(ctx *Context, b []byte) (err error) {
	if err := j.preElem(); err != nil {
		return err
	}
	_, err = base64.NewEncoder(base64.StdEncoding, j.Output).Write(b)
	if err != nil {
		return
	}
	j.incrementElemIndex()
	return
}

func (j *JsonWriter) PushInt(ctx *Context, i int64) error {
	if err := j.preElem(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(j.Output, "%d", i)
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) PushUint(ctx *Context, u uint64) error {
	if err := j.preElem(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(j.Output, "%d", u)
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) PushFloat(ctx *Context, f float64) error {
	if err := j.preElem(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(j.Output, "%f", f)
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) PushBool(ctx *Context, b bool) (err error) {
	if err := j.preElem(); err != nil {
		return err
	}
	if b {
		_, err = fmt.Fprintf(j.Output, "true")
	} else {
		_, err = fmt.Fprintf(j.Output, "false")
	}
	j.incrementElemIndex()
	return
}

func (j *JsonWriter) PushNull(*Context) error {
	_, err := fmt.Fprintf(j.Output, "null")
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) BeginArray(*Context) (int, error) {
	if err := j.preElem(); err != nil {
		return 0, err
	}
	j.elemIndex = append(j.elemIndex, 0)
	_, err := fmt.Fprintf(j.Output, "[")
	j.keyWritten = false
	return 0, err
}

func (j *JsonWriter) EndArray(*Context, int) error {
	j.elemIndex = j.elemIndex[:len(j.elemIndex)-1]
	_, err := fmt.Fprintf(j.Output, "]")
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) BeginObject(*Context) (int, error) {
	if err := j.preElem(); err != nil {
		return 0, err
	}
	j.elemIndex = append(j.elemIndex, 0)
	_, err := fmt.Fprintf(j.Output, "{")
	j.keyWritten = false
	return 0, err
}

func (j *JsonWriter) EndObject(*Context, int) error {
	j.elemIndex = j.elemIndex[:len(j.elemIndex)-1]
	_, err := fmt.Fprintf(j.Output, "}")
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) preElem() (err error) {
	if len(j.elemIndex) > 0 && j.elemIndex[len(j.elemIndex)-1] > 0 && !j.keyWritten {
		_, err = fmt.Fprintf(j.Output, ",")
	}
	return
}

func (j *JsonWriter) incrementElemIndex() {
	if len(j.elemIndex) > 0 {
		j.elemIndex[len(j.elemIndex)-1]++
	}
	j.keyWritten = false
}

func (j *JsonWriter) PushObjectKey(ctx *Context, k string) (err error) {
	if err := j.preElem(); err != nil {
		return err
	}
	buf := make([]byte, 0, len(k))
	buf = flexbuffers.EscapeJSONString(buf, k)
	_, err = j.Output.Write(buf)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(j.Output, ":")
	j.keyWritten = true
	return
}
