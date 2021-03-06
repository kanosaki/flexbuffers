package flexbuffers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtReadWrite(t *testing.T) {
	cases := []struct {
		name     string
		buildFn  func(b *Builder)
		assertFn func(a *assert.Assertions, r Raw)
	}{
		{
			name: "string with ext",
			buildFn: func(b *Builder) {
				b.Ext(123)
				b.StringValue("hello")
			},
			assertFn: func(a *assert.Assertions, r Raw) {
				root, err := r.Root()
				if !a.NoError(err) {
					return
				}
				sRef := root.AsStringRef()
				a.Equal("hello", sRef.StringValueOrEmpty())
				a.Equal(int64(123), sRef.Ext())
			},
		},
		{
			name: "string without ext",
			buildFn: func(b *Builder) {
				b.StringValue("hello")
			},
			assertFn: func(a *assert.Assertions, r Raw) {
				root, err := r.Root()
				if !a.NoError(err) {
					return
				}
				sRef := root.AsStringRef()
				a.Equal("hello", sRef.StringValueOrEmpty())
				a.Equal(int64(0), sRef.Ext())
			},
		},
		{
			name: "blob with ext",
			buildFn: func(b *Builder) {
				b.Ext(-123)
				b.Blob([]byte("world"))
			},
			assertFn: func(a *assert.Assertions, r Raw) {
				root, err := r.Root()
				if !a.NoError(err) {
					return
				}
				sRef := root.AsBlob()
				a.Equal([]byte("world"), sRef.DataOrEmpty())
				a.Equal(int64(-123), sRef.Ext())
			},
		},
		{
			name: "vector with ext",
			buildFn: func(b *Builder) {
				b.Ext(456)
				b.Vector(false, false, func(bld *Builder) {
					b.Int(1)
					b.Int(12345678)
				})
			},
			assertFn: func(a *assert.Assertions, r Raw) {
				root, err := r.Root()
				if !a.NoError(err) {
					return
				}
				v := root.AsVector()
				a.Equal(int64(1), v.AtOrNull(0).AsInt64())
				a.Equal(int64(456), v.Ext())
				a.Equal(int64(456), root.Ext())
			},
		},
		{
			name: "map with ext",
			buildFn: func(b *Builder) {
				b.Ext(-456)
				b.Map(func(bld *Builder) {
					b.IntField([]byte("a"), 1)
				})
			},
			assertFn: func(a *assert.Assertions, r Raw) {
				root, err := r.Root()
				if !a.NoError(err) {
					return
				}
				v := root.AsMap()
				a.Equal(int64(1), v.GetOrNull("a").AsInt64())
				a.Equal(int64(-456), v.Ext())
			},
		},
		{
			name: "map and its elements have ext",
			buildFn: func(b *Builder) {
				b.Ext(-456)
				b.Map(func(bld *Builder) {
					b.Ext(123)
					b.StringValueField([]byte("a"), "abc")
					b.Ext(456)
					b.BlobField([]byte("b"), []byte("foo"))
					b.Ext(789)
					b.VectorField([]byte("c"), false, false, func(bld *Builder) {
						b.Int(123)
					})
				})
			},
			assertFn: func(a *assert.Assertions, r Raw) {
				root, err := r.Root()
				if !a.NoError(err) {
					return
				}
				v := root.AsMap()
				a.Equal("abc", v.GetOrNull("a").AsStringRef().StringValueOrEmpty())
				a.Equal(int64(123), v.GetOrNull("a").Ext())
				a.Equal([]byte("foo"), v.GetOrNull("b").AsBlob().DataOrEmpty())
				a.Equal(int64(456), v.GetOrNull("b").Ext())
				a.Equal(int64(123), v.GetOrNull("c").AsVector().AtOrNull(0).AsInt64())
				a.Equal(int64(789), v.GetOrNull("c").Ext())
				a.Equal(int64(789), v.GetOrNull("c").AsVector().Ext())
				a.Equal(int64(-456), v.Ext())
			},
		},
	}
	for _, cas := range cases {
		t.Run(cas.name, func(t *testing.T) {
			b := NewBuilderWithFlags(BuilderFlagShareAll)
			cas.buildFn(b)
			if err := b.Finish(); err != nil {
				t.Fatal(err)
			}
			cas.assertFn(assert.New(t), b.Buffer())
		})
	}
}
