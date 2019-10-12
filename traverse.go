package flexbuffers

import (
	"sort"
	"unsafe"
)

type Traverser struct {
	buf         Raw
	offset      int
	typ         Type
	byteWidth   int
	parentWidth int
}

func (t *Traverser) digMap(key string) {
	mapOffset := t.buf.Indirect(t.offset, uint8(t.parentWidth))
	keyBytes := *(*[]byte)(unsafe.Pointer(&key))
	keysVectorInd := mapOffset - t.byteWidth*3
	keysByteWidth := int(t.buf.ReadUInt64(keysVectorInd+t.byteWidth, uint8(t.byteWidth)))
	keysOffset := t.buf.Indirect(keysVectorInd, uint8(t.byteWidth))
	keysLen := int(t.buf.ReadUInt64(keysOffset-keysByteWidth, uint8(keysByteWidth)))

	foundIdx := sort.Search(keysLen, func(i int) bool {
		ind := t.buf.Indirect(keysOffset+i*keysByteWidth, uint8(keysByteWidth))
		for i, c := range keyBytes {
			kc := t.buf[ind+i]
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
	if foundIdx < keysLen { // found
		keyDataOffset := t.buf.Indirect(keysOffset+foundIdx*keysByteWidth, uint8(keysByteWidth))
		exactEqual := true
		for i, c := range keyBytes {
			kc := t.buf[keyDataOffset+i]
			if kc == 0 || kc != c {
				exactEqual = false
				break
			}
		}
		if exactEqual {
			valuePackedType := t.buf[mapOffset+keysLen*t.byteWidth+foundIdx]
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
}

func (t *Traverser) Seek(path []string) {
	for _, p := range path {
		if t.typ != FBTMap {
			return
		}
		t.digMap(p)
	}
}

func (t *Traverser) Current() Reference {
	return Reference{
		data_:       t.buf,
		offset:      t.offset,
		parentWidth: uint8(t.parentWidth),
		byteWidth:   uint8(t.byteWidth),
		type_:       t.typ,
	}
}
