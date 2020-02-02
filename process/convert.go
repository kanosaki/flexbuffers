package process

type DocumentReader interface {
	SetOutput(w DocumentWriter) error
	ReadBuffer(b []byte) error
	// TODO: add function that accepts io.Reader as Input
}
type DocumentWriter interface {
	PushString(s string) error
	PushBlob(b []byte) error
	PushInt(i int64) error
	PushUint(u uint64) error
	PushFloat(f float64) error
	PushBool(b bool) error
	PushNull() error

	BeginArray() (int, error)
	EndArray(int) error

	BeginObject() (int, error)
	EndObject(int) error
	PushObjectKey(k string) error
}
