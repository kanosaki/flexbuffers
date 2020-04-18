package process

type Context struct {
	LastId int
}

func (c *Context) NextID() int {
	c.LastId++
	return c.LastId
}

type Document interface {
	FlushTo(ctx *Context, w DocumentWriter) error
}

type DocumentWriter interface {
	PushString(ctx *Context, s string) error
	PushBlob(ctx *Context, b []byte) error
	PushInt(ctx *Context, i int64) error
	PushUint(ctx *Context, u uint64) error
	PushFloat(ctx *Context, f float64) error
	PushBool(ctx *Context, b bool) error
	PushNull(ctx *Context) error

	BeginArray(ctx *Context) (int, error)
	EndArray(ctx *Context, id int) error

	BeginObject(ctx *Context) (int, error)
	EndObject(ctx *Context, id int) error
	PushObjectKey(ctx *Context, k string) error
}

type DocumentProcessor interface {
	DocumentWriter
	Reset(ctx *Context, w DocumentWriter)
}
