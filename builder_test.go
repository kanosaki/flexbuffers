package flexbuffers

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAndParse(t *testing.T) {
	cases := []struct {
		name     string
		buildFn  func(b *Builder)
		assertFn func(a *assert.Assertions, r Reference)
	}{
		{
			name: "simple_int16",
			buildFn: func(b *Builder) {
				b.Int(234)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal(2, int(r.parentWidth)) // use parentWidth because int is inline type
				a.Equal(int64(234), r.AsInt64())
			},
		},
		{
			name: "simple_int16_negative",
			buildFn: func(b *Builder) {
				b.Int(-234)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal(2, int(r.parentWidth)) // use parentWidth because int is inline type
				a.Equal(int64(-234), r.AsInt64())
			},
		},
		{
			name: "simple_uint16",
			buildFn: func(b *Builder) {
				b.UInt(math.MaxUint8 + 1)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal(2, int(r.parentWidth)) // use parentWidth because int is inline type
				a.Equal(uint64(math.MaxUint8+1), r.AsUInt64())
			},
		},
		{
			name: "simple_uint16_upper",
			buildFn: func(b *Builder) {
				b.UInt(math.MaxInt16 + 1)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal(2, int(r.parentWidth)) // use parentWidth because int is inline type
				a.Equal(uint64(math.MaxInt16+1), r.AsUInt64())
			},
		},
		{
			name: "simple_string",
			buildFn: func(b *Builder) {
				b.StringValue("hello world")
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal("hello world", r.AsStringRef().StringValue())
				a.Equal("hello world", r.AsStringRef().UnsafeStringValue())
			},
		},
		{
			name: "simple_blob",
			buildFn: func(b *Builder) {
				b.Blob([]byte{10, 20, 30})
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal([]byte{10, 20, 30}, r.AsBlob().Data())
			},
		},
		{
			name: "simple_null",
			buildFn: func(b *Builder) {
				b.Null()
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.True(r.IsNull())
			},
		},
		{
			name: "simple_vector",
			buildFn: func(b *Builder) {
				ptr := b.StartVector()
				b.Int(10)
				b.Int(20)
				b.Int(30)
				b.EndVector(ptr, false, false)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				vec := r.AsVector()
				a.Equal(int64(10), vec.At(0).AsInt64())
				a.Equal(int64(20), vec.At(1).AsInt64())
				a.Equal(int64(30), vec.At(2).AsInt64())
			},
		},
		{
			name: "flat_polymorphic_vector",
			buildFn: func(b *Builder) {
				ptr := b.StartVector()
				b.Int(10)
				b.UInt(20)
				b.StringValue("Hello")
				b.EndVector(ptr, false, false)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				vec := r.AsVector()
				a.Equal(int64(10), vec.At(0).AsInt64())
				a.Equal(uint64(20), vec.At(1).AsUInt64())
				a.Equal("Hello", vec.At(2).AsStringRef().StringValue())
			},
		},
		{
			name: "nested_vector",
			buildFn: func(b *Builder) {
				v1 := b.StartVector()
				b.Int(10)
				b.StringValue("Hello")
				v2 := b.StartVector()
				b.StringValue("World")
				b.Int(20)
				b.EndVector(v2, false, false)
				b.EndVector(v1, false, false)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				v1 := r.AsVector()
				a.Equal(int64(10), v1.At(0).AsInt64())
				a.Equal("Hello", v1.At(1).AsStringRef().StringValue())
				v2 := v1.At(2).AsVector()
				a.Equal("World", v2.At(0).AsStringRef().StringValue())
				a.Equal(int64(20), v2.At(1).AsInt64())
			},
		},
		{
			name: "simple_map",
			buildFn: func(b *Builder) {
				m := b.StartMap()
				b.IntField([]byte("a"), 10)
				b.IntField([]byte("b"), 20)
				b.IntField([]byte("c"), 30)
				b.EndMap(m)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				m := r.AsMap()
				a.Equal(int64(10), m.GetOrNull("a").AsInt64())
				a.Equal(int64(20), m.GetOrNull("b").AsInt64())
				a.Equal(int64(30), m.GetOrNull("c").AsInt64())
			},
		},
		{
			name: "flat_polymorphic_map",
			buildFn: func(b *Builder) {
				m := b.StartMap()
				b.UIntField([]byte("a"), 10)
				b.IntField([]byte("b"), 20)
				b.StringValueField([]byte("c"), "HELLO")
				b.BlobField([]byte("d"), []byte("WORLD"))
				b.Float32Field([]byte("e"), 12.3)
				b.EndMap(m)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				m := r.AsMap()
				a.Equal(uint64(10), m.GetOrNull("a").AsUInt64())
				a.Equal(int64(20), m.GetOrNull("b").AsInt64())
				a.Equal("HELLO", m.GetOrNull("c").AsStringRef().StringValue())
				a.Equal([]byte("WORLD"), m.GetOrNull("d").AsBlob().Data())
				a.Equal(float32(12.3), m.GetOrNull("e").AsFloat32())
			},
		},
		{
			name: "nested_map",
			buildFn: func(b *Builder) {
				m1 := b.StartMap()
				b.IntField([]byte("a"), 123)
				m2 := b.StartMapField([]byte("b"))
				b.StringValueField([]byte("c"), "world")
				b.EndMap(m2)
				b.Float32Field([]byte("d"), 12.3)
				b.EndMap(m1)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				m1 := r.AsMap()
				a.Equal(int64(123), m1.GetOrNull("a").AsInt64())
				m2 := m1.GetOrNull("b").AsMap()
				a.Equal("world", m2.GetOrNull("c").AsStringRef().StringValue())
				a.Equal(float32(12.3), m1.GetOrNull("d").AsFloat32())
			},
		},
		{
			name: "large_doc1",
			buildFn: func(b *Builder) {
				b.Map(func(b *Builder) {
					for i := 0; i < 100; i++ {
						b.MapField([]byte(fmt.Sprintf("map-%d", i)), func(b *Builder) {
							for j := 0; j < 100; j++ {
								b.StringValueField([]byte(fmt.Sprintf("key-%d", j)), fmt.Sprintf("v-%d-%d", i, j))
							}
						})
					}
				})
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				m1 := r.AsMap()
				a.Equal(
					"v-80-90",
					m1.GetOrNull("map-80").AsMap().
						GetOrNull("key-90").AsStringRef().
						StringValue())
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := assert.New(t)
			b := NewBuilder()
			c.buildFn(b)
			if err := b.Finish(); err != nil {
				t.Error(err)
			}
			c.assertFn(a, b.Buffer().Root())
		})
	}
}
