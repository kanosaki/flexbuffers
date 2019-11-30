package scanbench

import (
	"bufio"
	"os"
	"testing"

	"github.com/valyala/fastjson"

	"flexbuffers"
)

func scanJsons(b *testing.B) [][]byte {
	var docs [][]byte
	f, err := os.Open("data.json")
	if err != nil {
		b.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		docs = append(docs, []byte(scanner.Text()))
	}
	if scanner.Err() != nil {
		b.Fatal(scanner.Err())
	}
	return docs
}

func BenchmarkScanFlexbuffer(b *testing.B) {
	jsonDocs := scanJsons(b)
	var docs []flexbuffers.Raw
	for _, d := range jsonDocs {
		fbDoc, err := flexbuffers.FromJson(d)
		if err != nil {
			b.Fatal(err)
		}
		docs = append(docs, fbDoc)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range docs {
			_, err := d.Lookup("entities", "media")
			if err != nil && err != flexbuffers.ErrNotFound {
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
