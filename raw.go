package flexbuffers

import (
	"errors"
)

var (
	ErrUpdateDoesntFit = errors.New("update data doesn't fit")
)

type Raw []byte

func (b Raw) Validate() error {
	root, err := b.Root()
	if err != nil {
		return err
	}
	return root.Validate()
}

func (b Raw) RootOrNull() Reference {
	v, err := b.Root()
	if err != nil {
		return NullReference
	}
	return v
}

func (b Raw) Root() (Reference, error) {
	if len(b) <= 2 {
		return Reference{}, ErrInvalidData
	}
	byteWidth := b[len(b)-1]
	packedType := b[len(b)-2]
	rootOffset := len(b) - 2 - int(byteWidth)
	return NewReferenceFromPackedType(&b, rootOffset, byteWidth, packedType)
}

func (b Raw) InitTraverser(tv *Traverser) {
	byteWidth := b[len(b)-1]
	packedType := b[len(b)-2]
	rootOffset := len(b) - 2 - int(byteWidth)
	*tv = Traverser{
		buf:         &b,
		offset:      rootOffset,
		typ:         Type(packedType >> 2),
		parentWidth: int(byteWidth),
		byteWidth:   1 << (packedType & 3),
	}
}

func (b Raw) LookupOrNull(path ...string) Reference {
	r, err := b.Lookup(path...)
	if err != nil {
		return NullReference
	}
	return r
}

func (b Raw) Lookup(path ...string) (Reference, error) {
	var tv Traverser
	b.InitTraverser(&tv)
	if err := tv.Seek(path); err != nil {
		return Reference{}, err
	}
	return tv.Current()
}
