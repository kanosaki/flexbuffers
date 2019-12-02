package scanbench

import (
	"encoding/json"
	"testing"

	"github.com/valyala/fastjson"
	"go.mongodb.org/mongo-driver/bson"

	"flexbuffers"
	"flexbuffers/convert"
	"flexbuffers/pkg/fakedata"
)

func scanJsons(b *testing.B) [][]byte {
	count := 100
	docs := make([][]byte, 0, count)
	err := fakedata.Tweets(100, func(t *fakedata.Tweet) error {
		buf, err := json.Marshal(t)
		if err != nil {
			return err
		}
		docs = append(docs, buf)
		return nil
	})
	if err != nil {
		b.Fatal(err)
	}
	return docs
}

func BenchmarkScanBSON(b *testing.B) {
	count := 100
	bsonDocs := make([][]byte, 0, count)
	err := fakedata.Tweets(100, func(t *fakedata.Tweet) error {
		buf, err := bson.Marshal(t)
		if err != nil {
			return err
		}
		bsonDocs = append(bsonDocs, buf)
		return nil
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, docByte := range bsonDocs {
			d := bson.Raw(docByte)
			_, err := d.LookupErr("entities", "media")
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkScanFlexbuffer(b *testing.B) {
	jsonDocs := scanJsons(b)
	var docs []flexbuffers.Raw
	for _, d := range jsonDocs {
		fbDoc, err := convert.FromJson(d)
		if err != nil {
			b.Fatal(err)
		}
		docs = append(docs, fbDoc)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range docs {
			_, err := d.Lookup("entities", "media")
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkScanJson(b *testing.B) {
	docs := scanJsons(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range docs {
			v, err := fastjson.ParseBytes(d)
			if err != nil {
				b.Fatal(err)
			}
			_ = v.Get("entities", "media")
		}
	}
}
