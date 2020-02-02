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

func (j *JsonWriter) PushString(s string) (err error) {
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

func (j *JsonWriter) PushBlob(b []byte) (err error) {
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

func (j *JsonWriter) PushInt(i int64) error {
	if err := j.preElem(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(j.Output, "%d", i)
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) PushUint(u uint64) error {
	if err := j.preElem(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(j.Output, "%d", u)
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) PushFloat(f float64) error {
	if err := j.preElem(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(j.Output, "%f", f)
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) PushBool(b bool) (err error) {
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

func (j *JsonWriter) PushNull() error {
	_, err := fmt.Fprintf(j.Output, "null")
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) BeginArray() (int, error) {
	if err := j.preElem(); err != nil {
		return 0, err
	}
	j.elemIndex = append(j.elemIndex, 0)
	_, err := fmt.Fprintf(j.Output, "[")
	j.keyWritten = false
	return 0, err
}

func (j *JsonWriter) EndArray(int) error {
	j.elemIndex = j.elemIndex[:len(j.elemIndex)-1]
	_, err := fmt.Fprintf(j.Output, "]")
	j.incrementElemIndex()
	return err
}

func (j *JsonWriter) BeginObject() (int, error) {
	if err := j.preElem(); err != nil {
		return 0, err
	}
	j.elemIndex = append(j.elemIndex, 0)
	_, err := fmt.Fprintf(j.Output, "{")
	j.keyWritten = false
	return 0, err
}

func (j *JsonWriter) EndObject(int) error {
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

func (j *JsonWriter) PushObjectKey(k string) (err error) {
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
