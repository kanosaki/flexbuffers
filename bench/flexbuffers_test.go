package bench

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/valyala/fastjson"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/vmihailenco/msgpack"

	"flexbuffers"
)

var (
	objectKeys = []string{
		"a", "b", "c", "d", "e", "f", "g",
	}
)

func genMapField(childKeys []string, depth int) func(bld *flexbuffers.Builder) {
	if depth == 0 {
		return func(bld *flexbuffers.Builder) {}
	} else {
		return func(bld *flexbuffers.Builder) {
			for _, k := range childKeys {
				bld.MapField([]byte(k), genMapField(childKeys, depth-1))
			}
		}
	}
}
func BenchmarkFlexbuffersTraverseByReference(b *testing.B) {
	bld := flexbuffers.NewBuilder()
	bld.Map(genMapField(objectKeys, 5))
	if err := bld.Finish(); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root := bld.Buffer().Root()
		leaf := root.AsMap().Get("g").AsMap().Get("f").AsMap().Get("e").AsMap().Get("d").AsMap().Get("c").AsMap()
		if leaf.Size() != 0 {
			b.Fatal("assertion error")
		}
	}
}

func BenchmarkFlexbuffersTraverseByTraverser(b *testing.B) {
	bld := flexbuffers.NewBuilder()
	bld.Map(genMapField(objectKeys, 5))
	if err := bld.Finish(); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root := bld.Buffer()
		tv := root.Lookup("g", "f", "e", "d", "c")
		if tv.AsMap().Size() != 0 {
			b.Fatal("assertion error")
		}
	}
}

func genChildBson(keys []string, depth int) bson.D {
	if depth == 0 {
		return bson.D{}
	} else {
		var elems []bson.E
		for _, k := range keys {
			elems = append(elems, bson.E{
				Key:   k,
				Value: genChildBson(keys, depth-1),
			})
		}
		return elems
	}
}

func BenchmarkBSONTraverseTree(b *testing.B) {
	d := genChildBson(objectKeys, 5)
	data, err := bson.Marshal(d)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		raw := bson.Raw(data)
		e, err := raw.LookupErr("g", "f", "e", "d", "c")
		if err != nil {
			b.Fatal(err)
		}
		_, ok := e.DocumentOK()
		if !ok {
			b.Fatal("assertion error")
		}
	}
}

func genJsonObj(keys []string, depth int) interface{} {
	if depth == 0 {
		return make(map[string]interface{})
	} else {
		elems := make(map[string]interface{}, len(keys))
		for _, k := range keys {
			elems[k] = genJsonObj(keys, depth-1)
		}
		return elems
	}
}

func BenchmarkJSONTraverseByFastJson(b *testing.B) {
	obj := genJsonObj(objectKeys, 5)
	data, err := json.Marshal(obj)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, err := fastjson.ParseBytes(data)
		if err != nil {
			b.Fatal(err)
		}
		val := v.Get("g", "f", "e", "d", "c")
		if val == nil {
			b.Fatal("error")
		}
	}
}

func BenchmarkJSONTraverseByFastJsonWithoutParseTime(b *testing.B) {
	obj := genJsonObj(objectKeys, 5)
	data, err := json.Marshal(obj)
	if err != nil {
		b.Fatal(err)
	}
	v, err := fastjson.ParseBytes(data)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val := v.Get("g", "f", "e", "d", "c")
		if val == nil {
			b.Fatal("error")
		}
	}
}

func BenchmarkMsgpackTraverse(b *testing.B) {
	var buf bytes.Buffer
	err := msgpack.NewEncoder(&buf).Encode(genJsonObj(objectKeys, 5))
	if err != nil {
		b.Fatal(err)
	}
	data := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := msgpack.NewDecoder(bytes.NewBuffer(data))
		res, err := dec.Query("g.f.e.d.c")
		if err != nil {
			b.Fatal(err)
		}
		if len(res) == 0 {
			b.Fatal("empty")
		}
	}

}
