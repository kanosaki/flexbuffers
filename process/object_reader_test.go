package process

import (
	"math"
	"strings"
	"testing"

	"github.com/go-stack/stack"
	"github.com/google/go-cmp/cmp"
)

type PushRecord struct {
	Action string
	Params []interface{}
}

type MockWriter struct {
	History  []PushRecord
	ArrayCtr int
	MapCtr   int
}

func (m *MockWriter) Push(params ...interface{}) error {
	clr := stack.Caller(1)
	callerFunc := strings.Split(clr.Frame().Function, ".")
	m.History = append(m.History, PushRecord{
		Action: callerFunc[len(callerFunc)-1],
		Params: params,
	})
	return nil
}

func (m *MockWriter) PushString(ctx *Context, s string) error {
	return m.Push(s)
}

func (m *MockWriter) PushBlob(ctx *Context, b []byte) error {
	return m.Push(b)
}

func (m *MockWriter) PushInt(ctx *Context, i int64) error {
	return m.Push(i)
}

func (m *MockWriter) PushUint(ctx *Context, u uint64) error {
	return m.Push(u)
}

func (m *MockWriter) PushFloat(ctx *Context, f float64) error {
	return m.Push(f)
}

func (m *MockWriter) PushBool(ctx *Context, b bool) error {
	return m.Push(b)
}

func (m *MockWriter) PushNull(*Context) error {
	return m.Push()
}

func (m *MockWriter) BeginArray(*Context) (int, error) {
	m.ArrayCtr++
	return m.ArrayCtr, m.Push(m.ArrayCtr)
}

func (m *MockWriter) EndArray(ctx *Context, id int) error {
	return m.Push(id)
}

func (m *MockWriter) BeginObject(*Context) (int, error) {
	m.MapCtr++
	return m.MapCtr, m.Push(m.MapCtr)
}

func (m *MockWriter) EndObject(ctx *Context, id int) error {
	return m.Push(id)
}

func (m *MockWriter) PushObjectKey(ctx *Context, k string) error {
	return m.Push(k)
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
	mw := &MockWriter{}
	r := &ObjectReader{Output: mw}
	if err := r.Read(src); err != nil {
		t.Fatal(err)
	}
	expected := []PushRecord{
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
	if diff := cmp.Diff(expected, mw.History); diff != "" {
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
	mw := &MockWriter{}
	r := &ObjectReader{Output: mw}
	if err := r.Read(src); err != nil {
		t.Fatal(err)
	}
	expected := []PushRecord{
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
	if diff := cmp.Diff(expected, mw.History); diff != "" {
		t.Error(diff)
	}
}
