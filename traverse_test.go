package flexbuffers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraverser_Lookup(t *testing.T) {
	a := assert.New(t)
	b := NewBuilder()
	b.Map(func(b *Builder) {
		b.MapField([]byte("a"), func(b *Builder) {
			b.IntField([]byte("a"), 10)
			b.IntField([]byte("b"), 20)
		})
		b.MapField([]byte("b"), func(b *Builder) {
			b.IntField([]byte("c"), 30)
			b.IntField([]byte("d"), 40)
			b.StringValueField([]byte("e"), "foo")
		})
		b.IntField([]byte("c"), 123)
		b.MapField([]byte("d"), func(b *Builder) {
			b.MapField([]byte("a"), func(b *Builder) {
				b.StringValueField([]byte("b"), "bar")
			})
		})
	})
	if err := b.Finish(); err != nil {
		t.Fatal(err)
	}
	root := b.Buffer()
	a.Equal(int64(10), root.Lookup("a", "a").AsInt64())
	a.Equal(int64(40), root.Lookup("b", "d").AsInt64())
	a.Equal("foo", root.Lookup("b", "e").AsString().StringValue())
	a.Equal(int64(123), root.Lookup("c").AsInt64())
	a.True(root.Lookup("foo").IsNull())
	a.Equal("bar", root.Lookup("d", "a", "b").AsString().StringValue())
}

func TestTraverser_LookupLargeData(t *testing.T) {
	a := assert.New(t)
	b := NewBuilder()
	b.Map(func(b *Builder) {
		for i := 0; i < 100; i++ {
			b.MapField([]byte(fmt.Sprintf("a-%d", i)), func(b *Builder) {
				for j := 0; j < 100; j++ {
					b.IntField([]byte(fmt.Sprintf("b-%d", j)), int64(j)*10)
				}
			})
		}
	})
	if err := b.Finish(); err != nil {
		t.Fatal(err)
	}
	root := b.Buffer()
	a.Equal(int64(900), root.Lookup("a-50", "b-90").AsInt64())
}
