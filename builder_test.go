package flexbuffers

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAndParse(t *testing.T) {
	cases := []struct {
		name         string
		builderFlags BuilderFlag
		buildFn      func(b *Builder)
		assertFn     func(a *assert.Assertions, r Reference)
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
			name: "simple_float32",
			buildFn: func(b *Builder) {
				b.Float32(float32(math.Pi))
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal(4, int(r.parentWidth)) // use parentWidth because int is inline type
				a.Equal(float32(math.Pi), r.AsFloat32())
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
			name: "negative_number_boundaries",
			buildFn: func(b *Builder) {
				b.Vector(false, false, func(b *Builder) {
					b.Int(-1)
					b.Int(math.MinInt8)
					b.Int(math.MinInt16)
					b.Int(math.MinInt32)
					b.Int(math.MinInt64)
					b.Int(math.MinInt8 - 1)
					b.Int(math.MinInt16 - 1)
					b.Int(math.MinInt32 - 1)
				})
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				vec := r.AsVector()
				a.Equal(int64(-1), vec.AtOrNull(0).AsInt64())
				a.Equal(int64(math.MinInt8), vec.AtOrNull(1).AsInt64())
				a.Equal(int64(math.MinInt16), vec.AtOrNull(2).AsInt64())
				a.Equal(int64(math.MinInt32), vec.AtOrNull(3).AsInt64())
				a.Equal(int64(math.MinInt64), vec.AtOrNull(4).AsInt64())
				a.Equal(int64(math.MinInt8-1), vec.AtOrNull(5).AsInt64())
				a.Equal(int64(math.MinInt16-1), vec.AtOrNull(6).AsInt64())
				a.Equal(int64(math.MinInt32-1), vec.AtOrNull(7).AsInt64())
			},
		},
		{
			name: "simple_string",
			buildFn: func(b *Builder) {
				b.StringValue("hello world")
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal("hello world", r.AsStringRef().StringValueOrEmpty())
				a.Equal("hello world", r.AsStringRef().UnsafeStringValueOrEmpty())
			},
		},
		{
			name: "simple_blob",
			buildFn: func(b *Builder) {
				b.Blob([]byte{10, 20, 30})
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal([]byte{10, 20, 30}, r.AsBlob().DataOrEmpty())
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
				a.Equal(int64(10), vec.AtOrNull(0).AsInt64())
				a.Equal(int64(20), vec.AtOrNull(1).AsInt64())
				a.Equal(int64(30), vec.AtOrNull(2).AsInt64())
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
				a.Equal(int64(10), vec.AtOrNull(0).AsInt64())
				a.Equal(uint64(20), vec.AtOrNull(1).AsUInt64())
				a.Equal("Hello", vec.AtOrNull(2).AsStringRef().StringValueOrEmpty())
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
				a.Equal(int64(10), v1.AtOrNull(0).AsInt64())
				a.Equal("Hello", v1.AtOrNull(1).AsStringRef().StringValueOrEmpty())
				v2 := v1.AtOrNull(2).AsVector()
				a.Equal("World", v2.AtOrNull(0).AsStringRef().StringValueOrEmpty())
				a.Equal(int64(20), v2.AtOrNull(1).AsInt64())
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
			name: "unsorted_map",
			buildFn: func(b *Builder) {
				m := b.StartMap()
				b.IntField([]byte("b"), 20)
				b.IntField([]byte("c"), 30)
				b.IntField([]byte("a"), 10)
				b.EndMap(m)
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				m := r.AsMap()
				a.Equal(int64(10), m.GetOrNull("a").AsInt64())
				a.Equal(int64(20), m.GetOrNull("b").AsInt64())
				a.Equal(int64(30), m.GetOrNull("c").AsInt64())
				a.Equal(int64(10), m.AtOrNull(0).AsInt64())
				a.Equal(int64(20), m.AtOrNull(1).AsInt64())
				a.Equal(int64(30), m.AtOrNull(2).AsInt64())
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
				a.Equal("HELLO", m.GetOrNull("c").AsStringRef().StringValueOrEmpty())
				a.Equal([]byte("WORLD"), m.GetOrNull("d").AsBlob().DataOrEmpty())
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
				a.Equal("world", m2.GetOrNull("c").AsStringRef().StringValueOrEmpty())
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
						StringValueOrEmpty())
			},
		},
		{
			name:         "shared_key",
			builderFlags: BuilderFlagShareKeys,
			buildFn: func(b *Builder) {
				b.Vector(false, false, func(b *Builder) {
					b.Key([]byte("foo"))
					b.Key([]byte("foo"))
				})
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal("foo", r.AsVector().AtOrNull(0).AsKey().StringValue())
			},
		},
		{
			name:         "shared_key_and_key_vector",
			builderFlags: BuilderFlagShareKeys | BuilderFlagShareKeyVectors,
			buildFn: func(b *Builder) {
				b.Vector(false, false, func(b *Builder) {
					b.Map(func(b *Builder) {
						b.IntField([]byte("a"), 1)
						b.IntField([]byte("b"), 2)
					})
					b.Map(func(b *Builder) {
						b.IntField([]byte("a"), 3)
						b.IntField([]byte("b"), 4)
					})
				})
			},
			assertFn: func(a *assert.Assertions, r Reference) {
				a.Equal(int64(1), r.AsVector().AtOrNull(0).AsMap().GetOrNull("a").AsInt64())
				a.Equal(int64(2), r.AsVector().AtOrNull(0).AsMap().GetOrNull("b").AsInt64())
				a.Equal(int64(3), r.AsVector().AtOrNull(1).AsMap().GetOrNull("a").AsInt64())
				a.Equal(int64(4), r.AsVector().AtOrNull(1).AsMap().GetOrNull("b").AsInt64())
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := assert.New(t)
			b := NewBuilderWithFlags(c.builderFlags)
			c.buildFn(b)
			if err := b.Finish(); err != nil {
				t.Error(err)
			}
			c.assertFn(a, b.Buffer().RootOrNull())
		})
	}
}

func TestBuilder_KeyShare(t *testing.T) {
	a := assert.New(t)
	b := NewBuilderWithFlags(BuilderFlagShareKeys)
	b.Vector(false, false, func(b *Builder) {
		b.Key([]byte("a"))
		b.Key([]byte("b"))
		b.Key([]byte("a"))
	})
	if err := b.Finish(); err != nil {
		t.Fatal(err)
	}
	expected := []byte{
		'a', 0,
		'b', 0,
		3, // vector length
		5, // offset to 'a'
		4, // offset to 'b'
		7, // offset to 'a'
		PackedType(BitWidth8, FBTKey, false),
		PackedType(BitWidth8, FBTKey, false),
		PackedType(BitWidth8, FBTKey, false),
		6, // root
		PackedType(BitWidth8, FBTVector, false),
		1,
	}
	a.Equal(expected, []byte(b.Buffer()))
}

func TestBuilder_StringShare(t *testing.T) {
	a := assert.New(t)
	b := NewBuilderWithFlags(BuilderFlagShareStrings)
	b.Vector(false, false, func(b *Builder) {
		b.StringValue("a")
		b.StringValue("b")
		b.StringValue("a")
	})
	if err := b.Finish(); err != nil {
		t.Fatal(err)
	}
	expected := []byte{
		1, 'a', 0,
		1, 'b', 0,
		3, // vector length
		6, // offset to 'a'
		4, // offset to 'b'
		8, // offset to 'a'
		PackedType(BitWidth8, FBTString, false),
		PackedType(BitWidth8, FBTString, false),
		PackedType(BitWidth8, FBTString, false),
		6, // root
		PackedType(BitWidth8, FBTVector, false),
		1,
	}
	a.Equal(expected, []byte(b.Buffer()))
}

func TestBuilder_KeyVectorShare(t *testing.T) {
	a := assert.New(t)
	b := NewBuilderWithFlags(BuilderFlagShareKeyVectors)
	b.Vector(false, false, func(b *Builder) {
		b.Map(func(b *Builder) {
			b.IntField([]byte("a"), 1)
			b.IntField([]byte("b"), 2)
		})
		b.Map(func(b *Builder) {
			b.IntField([]byte("c"), 1)
			b.IntField([]byte("a"), 2)
		})
		b.Map(func(b *Builder) {
			b.IntField([]byte("b"), 1)
			b.IntField([]byte("a"), 2)
		})
	})
	if err := b.Finish(); err != nil {
		t.Fatal(err)
	}
	vec := b.Buffer().RootOrNull().AsVector()
	m0Keys, _ := vec.AtOrNull(0).AsMap().Keys()
	m1Keys, _ := vec.AtOrNull(1).AsMap().Keys()
	m2Keys, _ := vec.AtOrNull(2).AsMap().Keys()
	a.NotEqual(0, m0Keys.offset)
	a.NotEqual(0, m1Keys.offset)
	a.NotEqual(0, m2Keys.offset)
	a.Equal(m0Keys.offset, m2Keys.offset)
	a.NotEqual(m0Keys.offset, m1Keys.offset)
	a.NotEqual(m1Keys.offset, m2Keys.offset)
}

func TestBuilder_KeyAndKeyVectorShare(t *testing.T) {
	a := assert.New(t)
	b := NewBuilderWithFlags(BuilderFlagShareKeyVectors | BuilderFlagShareKeys)
	b.Vector(false, false, func(b *Builder) {
		b.Map(func(b *Builder) {
			b.IntField([]byte("a"), 1)
			b.IntField([]byte("b"), 2)
		})
		b.Map(func(b *Builder) {
			b.IntField([]byte("c"), 1)
			b.IntField([]byte("a"), 2)
		})
		b.Map(func(b *Builder) {
			b.IntField([]byte("b"), 1)
			b.IntField([]byte("a"), 2)
		})
	})
	if err := b.Finish(); err != nil {
		t.Fatal(err)
	}
	vec := b.Buffer().RootOrNull().AsVector()
	m0Keys, _ := vec.AtOrNull(0).AsMap().Keys()
	m1Keys, _ := vec.AtOrNull(1).AsMap().Keys()
	m2Keys, _ := vec.AtOrNull(2).AsMap().Keys()
	a.NotEqual(0, m0Keys.offset)
	a.NotEqual(0, m1Keys.offset)
	a.NotEqual(0, m2Keys.offset)
	a.Equal(m0Keys.offset, m2Keys.offset)
	a.NotEqual(m0Keys.offset, m1Keys.offset)
	a.NotEqual(m1Keys.offset, m2Keys.offset)

	// doesn't share keys vector but shares each key data
	m00, _ := m0Keys.AtOrNull(0).indirect() // 'a' key
	m11, _ := m1Keys.AtOrNull(0).indirect() // 'a' key, note map at index 1 will be sorted
	a.Equal(m00, m11)
}
