package manipulation

import (
	"errors"

	"flexbuffers/process"
)

func Replace(with process.Document) process.Manipulation {
	return &replace{Value: with}
}

type replace struct {
	Value  process.Document
	target process.DocumentWriter

	used              bool
	replacingObjectId *int
	replacingArrayId  *int
}

func (r *replace) PushString(ctx *process.Context, s string) error {
	return r.flushValue(ctx)
}

func (r *replace) PushBlob(ctx *process.Context, b []byte) error {
	return r.flushValue(ctx)
}

func (r *replace) PushInt(ctx *process.Context, i int64) error {
	return r.flushValue(ctx)
}

func (r *replace) PushUint(ctx *process.Context, u uint64) error {
	return r.flushValue(ctx)
}

func (r *replace) PushFloat(ctx *process.Context, f float64) error {
	return r.flushValue(ctx)
}

func (r *replace) PushBool(ctx *process.Context, b bool) error {
	return r.flushValue(ctx)
}

func (r *replace) PushNull(ctx *process.Context) error {
	return r.flushValue(ctx)
}

func (r *replace) BeginArray(ctx *process.Context) (int, error) {
	if r.used || r.replacingObjectId != nil {
		// pass
		return ctx.NextID(), nil
	}
	r.used = true
	id := ctx.NextID()
	r.replacingArrayId = &id
	return id, nil
}

func (r *replace) EndArray(ctx *process.Context, id int) error {
	if r.replacingArrayId == nil {
		if r.replacingObjectId != nil {
			return nil
		}
		return errors.New("invalid state: no BeginArray but got EndArray")
	}
	if *r.replacingArrayId != id {
		return nil // pass
	}
	return r.flushValue(ctx)
}

func (r *replace) BeginObject(ctx *process.Context) (int, error) {
	if r.used || r.replacingArrayId != nil {
		// pass
		return ctx.NextID(), nil
	}
	r.used = true
	id := ctx.NextID()
	r.replacingObjectId = &id
	return id, nil
}

func (r *replace) EndObject(ctx *process.Context, id int) error {
	if r.replacingObjectId == nil {
		if r.replacingArrayId != nil {
			return nil
		}
		return errors.New("invalid state: no BeginObject but got EndObject")
	}
	if *r.replacingObjectId != id {
		return nil // pass
	}
	return r.flushValue(ctx)
}

func (r *replace) PushObjectKey(ctx *process.Context, k string) error {
	return nil
}

func (r *replace) flushValue(ctx *process.Context) error {
	if r.used {
		return errors.New("duplicate replace: please call Reset before perform replace again")
	}
	r.used = true
	return r.Value.FlushTo(ctx, r.target)
}

func (r replace) Reset(ctx *process.Context, w process.DocumentWriter) {
	r.replacingArrayId = nil
	r.replacingObjectId = nil
	r.target = w
}
