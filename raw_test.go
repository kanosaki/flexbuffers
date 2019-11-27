package flexbuffers

import (
	"io/ioutil"
	"math"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawNumberPrimitives(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		asserter func(r Reference, a *assert.Assertions)
	}{
		{
			"single number",
			[]byte{01, 04, 01},
			func(r Reference, a *assert.Assertions) {
				a.Equal(int64(1), r.AsInt64())
			},
		},
		{
			"single float32",
			[]byte{0, 0, 0x80, 0x3f, 0x0e, 0x04},
			func(r Reference, a *assert.Assertions) {
				a.Equal(float32(1.0), r.AsFloat32())
			},
		},
		// primitive with static_cast
		{
			"single float32 as float64",
			[]byte{0, 0, 0x80, 0x3f, 0x0e, 0x04},
			func(r Reference, a *assert.Assertions) {
				a.Equal(1.0, r.AsFloat64())
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := Raw(c.data)
			a := assert.New(t)
			root := r.RootOrNull()
			c.asserter(root, a)
		})
	}
}

func TestRawFromTestData(t *testing.T) {
	testdatadir := "testdata"
	cases := []struct {
		datafile string
		asserter func(r Reference, a *assert.Assertions)
	}{
		{
			"single_int_1.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal(int64(1), r.AsInt64())
				a.Equal(uint64(1), r.AsUInt64())
				a.Equal(float32(1.0), r.AsFloat32())
				a.Equal(1.0, r.AsFloat64())
			},
		},
		{
			"single_uint_1.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal(int64(1), r.AsInt64())
				a.Equal(uint64(1), r.AsUInt64())
				a.Equal(float32(1.0), r.AsFloat32())
				a.Equal(1.0, r.AsFloat64())
			},
		},
		{
			"single_float_1.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal(int64(1), r.AsInt64())
				a.Equal(uint64(1), r.AsUInt64())
				a.Equal(float32(1.0), r.AsFloat32())
				a.Equal(1.0, r.AsFloat64())
			},
		},
		{
			"single_double_1.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal(int64(1), r.AsInt64())
				a.Equal(uint64(1), r.AsUInt64())
				a.Equal(float32(1.0), r.AsFloat32())
				a.Equal(1.0, r.AsFloat64())
			},
		},
		{
			"primitive_corners.flexbuf",
			func(r Reference, a *assert.Assertions) {
				m := r.AsMap()
				a.Equal(int64(math.MaxInt32), m.GetOrNull("int32_max").AsInt64())
				a.Equal(int64(math.MinInt32), m.GetOrNull("int32_min").AsInt64())
				a.Equal(int64(math.MaxInt64), m.GetOrNull("int64_max").AsInt64())
				a.Equal(int64(math.MinInt64), m.GetOrNull("int64_min").AsInt64())
			},
		},
		{
			"single_indirect_int_1.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal(int64(1), r.AsInt64())
				a.Equal(uint64(1), r.AsUInt64())
				a.Equal(float32(1.0), r.AsFloat32())
				a.Equal(1.0, r.AsFloat64())
			},
		},
		{
			"single_indirect_float_1.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal(int64(1), r.AsInt64())
				a.Equal(uint64(1), r.AsUInt64())
				a.Equal(float32(1.0), r.AsFloat32())
				a.Equal(1.0, r.AsFloat64())
			},
		},
		{
			"single_indirect_double_1.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal(int64(1), r.AsInt64())
				a.Equal(uint64(1), r.AsUInt64())
				a.Equal(float32(1.0), r.AsFloat32())
				a.Equal(1.0, r.AsFloat64())
			},
		},
		{
			"simple_string.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal("hello flexbuffers!", r.AsStringRef().StringValueOrEmpty())
			},
		},
		{
			"simple_blob.flexbuf",
			func(r Reference, a *assert.Assertions) {
				a.Equal([]byte{0, 3, 9, 0, 0}, r.AsBlob().DataOrEmpty())
			},
		},
		{
			"simple_vector.flexbuf",
			func(r Reference, a *assert.Assertions) {
				v := r.AsVector()
				a.Equal(int64(1), v.AtOrNull(0).AsInt64())
				a.Equal(int64(256), v.AtOrNull(1).AsInt64())
				a.Equal(int64(65546), v.AtOrNull(2).AsInt64())
			},
		},
		{
			"simple_typed_vector.flexbuf",
			func(r Reference, a *assert.Assertions) {
				v := r.AsTypedVector()
				a.Equal(int64(1), v.AtOrNull(0).AsInt64())
				a.Equal(int64(256), v.AtOrNull(1).AsInt64())
				a.Equal(int64(65546), v.AtOrNull(2).AsInt64())
			},
		},
		{
			"simple_fixed_typed_vector.flexbuf",
			func(r Reference, a *assert.Assertions) {
				v := r.AsFixedTypedVector()
				a.Equal(int64(1), v.AtOrNull(0).AsInt64())
				a.Equal(int64(256), v.AtOrNull(1).AsInt64())
				a.Equal(int64(65546), v.AtOrNull(2).AsInt64())
			},
		},
		{
			"simple_map.flexbuf",
			func(r Reference, a *assert.Assertions) {
				m := r.AsMap()
				s := m.GetOrNull("foo").AsStringRef()
				a.Equal(1, m.SizeOrZero())
				a.Equal("bar", s.StringValueOrEmpty())
				a.Equal("bar", s.UnsafeStringValueOrEmpty())
			},
		},
		{
			"flat_multiple_map.flexbuf",
			func(r Reference, a *assert.Assertions) {
				m := r.AsMap()
				a.Equal(3, m.SizeOrZero())
				s := m.GetOrNull("foo").AsStringRef()
				a.Equal("bar", s.StringValueOrEmpty())
				a.Equal("bar", s.UnsafeStringValueOrEmpty())
				i := m.GetOrNull("a").AsInt64()
				a.Equal(int64(123), i)
				d := m.GetOrNull("b").AsFloat64()
				a.Equal(12.0, d)
			},
		},
		{
			"nested_map_vector.flexbuf",
			func(r Reference, a *assert.Assertions) {
				m := r.AsMap()
				a.Equal(3, m.SizeOrZero())
				a.Equal(int64(123), m.GetOrNull("int").AsInt64())
				m1 := m.GetOrNull("map").AsMap()
				a.Equal(1, m1.SizeOrZero())
				a.Equal("bar", m1.GetOrNull("foo").AsStringRef().StringValueOrEmpty())
				v1 := m.GetOrNull("vec").AsVector()
				a.Equal(3, v1.SizeOrZero())
				a.Equal(int64(1), v1.AtOrNull(0).AsInt64())
				a.Equal(int64(256), v1.AtOrNull(1).AsInt64())
				a.Equal(int64(65546), v1.AtOrNull(2).AsInt64())
			},
		},
	}
	for _, c := range cases {
		t.Run(c.datafile, func(t *testing.T) {
			datafile := path.Join(testdatadir, c.datafile)
			data, err := ioutil.ReadFile(datafile)
			if err != nil {
				t.Fatal(err)
			}
			r := Raw(data)
			a := assert.New(t)
			root := r.RootOrNull()
			c.asserter(root, a)
		})
	}
}

func TestRawOffsetError(t *testing.T) {
	a := assert.New(t)
	r := Raw([]byte{})
	for i := 0; i < 4; i++ {
		bw := uint8(math.Pow(2, float64(i)))
		_, err := r.ReadInt64(0, bw)
		a.Equal(ErrOutOfRange, err)
		_, err = r.ReadUInt64(0, bw)
		a.Equal(ErrOutOfRange, err)
		_, err = r.ReadDouble(0, bw)
		a.Equal(ErrOutOfRange, err)
	}
	r = []byte{1, 2, 3}
	for i := 0; i < 4; i++ {
		bw := uint8(math.Pow(2, float64(i)))
		_, err := r.ReadInt64(2, bw)
		a.Equal(ErrOutOfRange, err)
		_, err = r.ReadUInt64(2, bw)
		a.Equal(ErrOutOfRange, err)
		_, err = r.ReadDouble(2, bw)
		a.Equal(ErrOutOfRange, err)
	}
}
