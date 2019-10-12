package flexbuffers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue_AsBool(t *testing.T) {
	a := assert.New(t)
	v := newValueBool(true)
	a.Equal(value{
		d:           1,
		typ:         FBTBool,
		minBitWidth: 0,
	}, v)
	v = newValueBool(false)
	a.Equal(value{
		d:           0,
		typ:         FBTBool,
		minBitWidth: 0,
	}, v)
}

func TestValue_AsInt(t *testing.T) {
	a := assert.New(t)
	vs := []int64{
		0, -1,
		math.MaxInt8, math.MaxInt16, math.MaxInt32, math.MaxInt64,
		math.MinInt8, math.MinInt16, math.MinInt32, math.MinInt64,
	}
	for _, v := range vs {
		val := newValueInt(v, FBTInt, BitWidth64)
		a.Equal(v, val.AsInt())
	}
}

func TestValue_AsUInt(t *testing.T) {
	a := assert.New(t)
	vs := []uint64{
		0,
		math.MaxUint8, math.MaxUint16, math.MaxUint32, math.MaxUint64,
	}
	for _, v := range vs {
		val := newValueUInt(v, FBTUint, BitWidth64)
		a.Equal(v, val.AsUInt())
	}
}

func TestValue_AsFloat(t *testing.T) {
	a := assert.New(t)
	vs := []float64{
		0,
		0.1,
		math.Pi,
		math.E,
		math.MaxFloat32, math.MaxFloat64,
		math.SmallestNonzeroFloat32, math.SmallestNonzeroFloat64,
	}
	for _, v := range vs {
		val := newValueFloat64(v)
		a.Equal(v, val.AsFloat())
	}
}

//func TestValue_ElemWidth(t *testing.T) {
//	a := assert.New(t)
//	cases := []struct {
//		v        value
//		expected BitWidth
//	}{
//		{
//			value{
//				d:           0,
//				typ:         0,
//				minBitWidth: 0,
//			},
//			BitWidth64,
//		},
//	}
//	for _, c := range cases {
//		a.Equal(c.expected, c.v.ElemWidth(0, 0))
//	}
//}

//func TestValue_StoredWidth(t *testing.T) {
//	a := assert.New(t)
//	cases := []struct {
//		v        value
//		parent   BitWidth
//		expected BitWidth
//	}{
//		{
//			value{
//				d:           0,
//				typ:         0,
//				minBitWidth: 0,
//			},
//			BitWidth8,
//			BitWidth64,
//		},
//	}
//	for _, c := range cases {
//		a.Equal(c.expected, c.v.ElemWidth(0, 0))
//	}
//}
