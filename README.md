## FlexBuffers


FlexBuffers is a schema-less, dynamic format optimized for reading single value in a document. It's a child format of FlatBuffers.

For more detailed specification, please refer <https://google.github.io/flatbuffers/flexbuffers.html>

NOTE: I'm planning extend flexbuffers by adding ability to attach metadata to each element. 
And I'm going to rename project name as it will lose few backward compatibility. (type bit space is too tight to attach metadata information, so FBT_VECTOR_BOOL might be no longer available in forked version)
