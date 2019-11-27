package flexbuffers

import (
	"errors"
	"sort"
	"unsafe"
)

var (
	ErrNotFound = errors.New("no element")
)

// Traverser provides optimized document traversing. Use Raw.Lookup if there is no special reason.
type Traverser struct {
	buf         *Raw
	offset      int
	typ         Type
	byteWidth   int
	parentWidth int
}

func (t *Traverser) digMap(key string) error {
	mapOffset, err := t.buf.Indirect(t.offset, uint8(t.parentWidth))
	if err != nil {
		return err
	}
	keyBytes := *(*[]byte)(unsafe.Pointer(&key))
	keysVectorInd := mapOffset - t.byteWidth*3
	keysByteWidth64, err := t.buf.ReadUInt64(keysVectorInd+t.byteWidth, uint8(t.byteWidth))
	if err != nil {
		return err
	}
	keysByteWidth := int(keysByteWidth64)
	keysOffset, err := t.buf.Indirect(keysVectorInd, uint8(t.byteWidth))
	if err != nil {
		return err
	}
	keysLen64, err := t.buf.ReadUInt64(keysOffset-keysByteWidth, uint8(keysByteWidth))
	if err != nil {
		return err
	}
	keysLen := int(keysLen64)

	var searchErr error
	foundIdx := sort.Search(keysLen, func(i int) bool {
		ind, err := t.buf.Indirect(keysOffset+i*keysByteWidth, uint8(keysByteWidth))
		if err != nil {
			searchErr = err
			return true
		}
		for i, c := range keyBytes {
			kc := (*t.buf)[ind+i]
			if kc == 0 {
				return false // -1
			} else if kc > c {
				return true //1
			} else if kc < c {
				return false //  -1
			}
		}
		return true
	})
	if searchErr != nil {
		return searchErr
	}
	if foundIdx < keysLen { // found
		keyDataOffset, err := t.buf.Indirect(keysOffset+foundIdx*keysByteWidth, uint8(keysByteWidth))
		if err != nil {
			return err
		}
		exactEqual := true
		for i, c := range keyBytes {
			kc := (*t.buf)[keyDataOffset+i]
			if kc == 0 || kc != c {
				exactEqual = false
				break
			}
		}
		if exactEqual {
			valuePackedType := (*t.buf)[mapOffset+keysLen*t.byteWidth+foundIdx]
			valueOffset := mapOffset + foundIdx*t.byteWidth

			// proceed
			t.parentWidth = t.byteWidth
			t.offset = valueOffset
			t.typ = Type(valuePackedType >> 2)
			t.byteWidth = 1 << (valuePackedType & 3)
		} else {
			t.offset = -1
			t.typ = FBTNull
		}
	} else {
		t.offset = -1
		t.typ = FBTNull
	}
	return nil
}

func (t *Traverser) Seek(path []string) error {
	for _, p := range path {
		if t.typ != FBTMap {
			return ErrNotFound
		}
		if err := t.digMap(p); err != nil {
			return err
		}
	}
	return nil
}

func (t *Traverser) Current() (Reference, error) {
	if t.offset == -1 {
		return Reference{}, ErrNotFound
	}
	r := Reference{
		data_:       t.buf,
		offset:      t.offset,
		parentWidth: uint8(t.parentWidth),
		byteWidth:   uint8(t.byteWidth),
		type_:       t.typ,
	}
	return r, r.CheckBoundary()
}
