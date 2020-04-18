package process

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"flexbuffers/process/manipulation"
)

func TestManipulationTreeNode(t *testing.T) {
	output := map[string]interface{}{
		"a": 123,
		"b": "456",
	}
	input := map[string]interface{}{
		"a": 123,
		"b": "456",
	}
	expected := map[string]interface{}{
		"a": "abc",
		"b": "456",
	}
	w, err := NewObjectWriter(output)
	if err != nil {
		t.Fatal(err)
	}
	man := NewManipulator(w)
	man.AddManipulation([]string{"a"}, manipulation.Replace("abc"))
	r := &ObjectReader{man}
	if err := r.Read(input); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(expected, output); diff != "" {
		t.Fatalf(diff)
	}
}
