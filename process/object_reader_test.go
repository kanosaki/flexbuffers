package process

import (
	"math"
	"strings"
	"testing"

	"github.com/go-stack/stack"
	"github.com/google/go-cmp/cmp"
)

type pushRecord struct {
	Action string
	Params []interface{}
}

type mockWriter struct {
	history  []pushRecord
	arrayCtr int
	mapCtr   int
}

func (m *mockWriter) push(params ...interface{}) error {
	clr := stack.Caller(1)
	callerFunc := strings.Split(clr.Frame().Function, ".")
	m.history = append(m.history, pushRecord{
		Action: callerFunc[len(callerFunc)-1],
		Params: params,
	})
	return nil
}

func (m *mockWriter) PushString(s string) error {
	return m.push(s)
}

func (m *mockWriter) PushBlob(b []byte) error {
	return m.push(b)
}

func (m *mockWriter) PushInt(i int64) error {
	return m.push(i)
}

func (m *mockWriter) PushUint(u uint64) error {
	return m.push(u)
}

func (m *mockWriter) PushFloat(f float64) error {
	return m.push(f)
}

func (m *mockWriter) PushBool(b bool) error {
	return m.push(b)
}

func (m *mockWriter) PushNull() error {
	return m.push()
}

func (m *mockWriter) BeginArray() (int, error) {
	m.arrayCtr++
	return m.arrayCtr, m.push(m.arrayCtr)
}

func (m *mockWriter) EndArray(i int) error {
	return m.push(i)
}

func (m *mockWriter) BeginObject() (int, error) {
	m.mapCtr++
	return m.mapCtr, m.push(m.mapCtr)
}

func (m *mockWriter) EndObject(i int) error {
	return m.push(i)
}

func (m *mockWriter) PushObjectKey(k string) error {
	return m.push(k)
}

func TestReadObject(t *testing.T) {
	type testObj struct {
		A int
		B string
		C []int
		D map[string]int
		E float64
	}
	src := &testObj{
		A: 1,
		B: "foo",
		C: []int{1, 2, 3},
		D: map[string]int{"x": 100},
		E: math.Pi,
	}
	mw := &mockWriter{}
	r := &ObjectReader{Output: mw}
	if err := r.Read(src); err != nil {
		t.Fatal(err)
	}
	expected := []pushRecord{
		{Action: "BeginObject", Params: []interface{}{1}},
		{Action: "PushObjectKey", Params: []interface{}{"A"}},
		{Action: "PushInt", Params: []interface{}{int64(1)}},
		{Action: "PushObjectKey", Params: []interface{}{"B"}},
		{Action: "PushString", Params: []interface{}{"foo"}},
		{Action: "PushObjectKey", Params: []interface{}{"C"}},
		{Action: "BeginArray", Params: []interface{}{1}},
		{Action: "PushInt", Params: []interface{}{int64(1)}},
		{Action: "PushInt", Params: []interface{}{int64(2)}},
		{Action: "PushInt", Params: []interface{}{int64(3)}},
		{Action: "EndArray", Params: []interface{}{1}},
		{Action: "PushObjectKey", Params: []interface{}{"D"}},
		{Action: "BeginObject", Params: []interface{}{2}},
		{Action: "PushObjectKey", Params: []interface{}{"x"}},
		{Action: "PushInt", Params: []interface{}{int64(100)}},
		{Action: "EndObject", Params: []interface{}{2}},
		{Action: "PushObjectKey", Params: []interface{}{"E"}},
		{Action: "PushFloat", Params: []interface{}{math.Pi}},
		{Action: "EndObject", Params: []interface{}{1}},
	}
	if diff := cmp.Diff(expected, mw.history); diff != "" {
		t.Error(diff)
	}
}

func TestReadObjectNested(t *testing.T) {
	type testObjInner struct {
		X string
	}
	type testObjOuter struct {
		A int
		B *testObjInner
		C []*testObjInner
		D []testObjInner
	}
	src := &testObjOuter{
		A: 1,
		B: &testObjInner{
			X: "abc",
		},
		C: []*testObjInner{
			{X: "1"},
			{X: "2"},
		},
		D: []testObjInner{
			{X: "a"},
			{X: "b"},
		},
	}
	mw := &mockWriter{}
	r := &ObjectReader{Output: mw}
	if err := r.Read(src); err != nil {
		t.Fatal(err)
	}
	expected := []pushRecord{
		{Action: "BeginObject", Params: []interface{}{1}},

		{Action: "PushObjectKey", Params: []interface{}{"A"}},
		{Action: "PushInt", Params: []interface{}{int64(1)}},

		{Action: "PushObjectKey", Params: []interface{}{"B"}},
		{Action: "BeginObject", Params: []interface{}{2}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushString", Params: []interface{}{"abc"}},
		{Action: "EndObject", Params: []interface{}{2}},

		{Action: "PushObjectKey", Params: []interface{}{"C"}},
		{Action: "BeginArray", Params: []interface{}{1}},
		{Action: "BeginObject", Params: []interface{}{3}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushString", Params: []interface{}{"1"}},
		{Action: "EndObject", Params: []interface{}{3}},
		{Action: "BeginObject", Params: []interface{}{4}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushString", Params: []interface{}{"2"}},
		{Action: "EndObject", Params: []interface{}{4}},
		{Action: "EndArray", Params: []interface{}{1}},

		{Action: "PushObjectKey", Params: []interface{}{"D"}},
		{Action: "BeginArray", Params: []interface{}{2}},
		{Action: "BeginObject", Params: []interface{}{5}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushString", Params: []interface{}{"a"}},
		{Action: "EndObject", Params: []interface{}{5}},
		{Action: "BeginObject", Params: []interface{}{6}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushString", Params: []interface{}{"b"}},
		{Action: "EndObject", Params: []interface{}{6}},
		{Action: "EndArray", Params: []interface{}{2}},

		{Action: "EndObject", Params: []interface{}{1}},
	}
	if diff := cmp.Diff(expected, mw.history); diff != "" {
		t.Error(diff)
	}
}
