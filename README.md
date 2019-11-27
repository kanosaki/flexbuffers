## FlexBuffers


FlexBuffers is a schema-less, dynamic format optimized for reading single value in a document. It's a child format of FlatBuffers.

For more detailed specification, please refer <https://google.github.io/flatbuffers/flexbuffers.html>

NOTE: I'm planning extend flexbuffers by adding ability to attach metadata to each element. 
And I'm going to rename project name as it will lose few backward compatibility. 
(type bit space is too tight to attach metadata information, so FBT_VECTOR_BOOL might be no longer available in forked version)


## Performance

Flexbuffers is optimized for lookup single value in a document. 
About 3x faster than BSON's `LookupErr` method. (`BenchmarkFlexbuffersTraverseByTraverser-4` vs `BenchmarkBSONTraverseTree-4`)



```bash
cd bench
go test -tags unsafe -bench . -benchmem
```

```
BenchmarkFlexbuffersTraverseByReference-4           	 1636011	       726 ns/op	       0 B/op	       0 allocs/op
BenchmarkFlexbuffersTraverseByTraverser-4           	 4406778	       272 ns/op	       0 B/op	       0 allocs/op
BenchmarkBSONTraverseTree-4                         	 1271134	       932 ns/op	       0 B/op	       0 allocs/op
BenchmarkJSONTraverseByFastJson-4                   	     277	   5343157 ns/op	10257193 B/op	   11229 allocs/op
BenchmarkJSONTraverseByFastJsonWithoutParseTime-4   	 6981626	       151 ns/op	       0 B/op	       0 allocs/op
BenchmarkMsgpackTraverse-4                          	    1900	    848233 ns/op	     288 B/op	       5 allocs/op
```