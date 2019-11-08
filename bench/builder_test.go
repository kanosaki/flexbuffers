package bench

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"flexbuffers"
)

func BenchmarkBuildMinimalDoc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bld := flexbuffers.NewBuilder()
		bld.Int(1)
		if err := bld.Finish(); err != nil {
			b.Fatal(err)
		}
	}
}

func buildTestDataFlexbufferRaw() ([]byte, error) {
	b := flexbuffers.NewBuilder()
	m0 := b.StartMap()
	m1 := b.StartMapField([]byte("x"))
	b.IntField([]byte("a"), 10)
	b.Float64Field([]byte("b"), math.E)
	b.StringValueField([]byte("c"), "hello")
	b.EndMap(m1)
	v1 := b.StartVectorField([]byte("y"))
	b.StringValue("foobar")
	b.Blob([]byte("abcd"))
	b.EndVector(v1, false, false)
	b.EndMap(m0)
	if err := b.Finish(); err != nil {
		return nil, err
	}
	return b.Buffer(), nil
}

func TestCheckBenchmarkFlexbufferRaw(t *testing.T) {
	res, err := buildTestDataFlexbufferRaw()
	if err != nil {
		t.Fatal(err)
	}
	fb := flexbuffers.Raw(res)
	a := assert.New(t)

	a.Equal(int64(10), fb.LookupOrNull("x", "a").AsInt64())
	fmt.Printf("Flexbuffers Size: %d\n", len(res))
}

func BenchmarkBuildSmallMapFlexbuffers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := buildTestDataFlexbufferRaw()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func buildTestDataBSON() ([]byte, error) {
	var err error
	raw := bson.Raw(nil)
	m0, raw := bsoncore.AppendDocumentStart(raw)
	m1, raw := bsoncore.AppendDocumentElementStart(raw, "x")
	raw = bsoncore.AppendInt32Element(raw, "a", 10)
	raw = bsoncore.AppendDoubleElement(raw, "b", math.E)
	raw = bsoncore.AppendStringElement(raw, "c", "hello")
	raw, err = bsoncore.AppendDocumentEnd(raw, m1)
	if err != nil {
		return nil, err
	}
	m2, raw := bsoncore.AppendDocumentElementStart(raw, "y")
	raw = bsoncore.AppendString(raw, "foobar")
	raw = bsoncore.AppendBinary(raw, 0, []byte("abcd"))
	raw, err = bsoncore.AppendDocumentEnd(raw, m2)
	if err != nil {
		return nil, err
	}
	return bsoncore.AppendDocumentEnd(raw, m0)
}

func TestCheckBenchmarkBSON(t *testing.T) {
	res, err := buildTestDataBSON()
	if err != nil {
		t.Fatal(err)
	}
	d := bsoncore.Document(res)
	a := assert.New(t)
	a.Equal(int32(10), d.Lookup("x", "a").Int32())
	fmt.Printf("BSON Size: %d\n", len(res))
}

func BenchmarkBuildSmallMapBSONRaw(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := buildTestDataBSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}
