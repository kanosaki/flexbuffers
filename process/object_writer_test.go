package process

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestObjectWriter(t *testing.T) {
	type innerStruct struct {
		X int
	}
	type dummyStruct struct {
		A int
		B string
		C map[string]interface{}
		D innerStruct
		E *innerStruct
		F []int
		G []*innerStruct
		H []innerStruct
		I *int
		J **int
	}
	target := &dummyStruct{}
	ow, err := NewObjectWriter(target)
	if err != nil {
		t.Fatal(err)
	}
	records := []PushRecord{
		{Action: "BeginObject", Params: []interface{}{}},

		{Action: "PushObjectKey", Params: []interface{}{"A"}},
		{Action: "PushInt", Params: []interface{}{int64(123)}},

		{Action: "PushObjectKey", Params: []interface{}{"B"}},
		{Action: "PushString", Params: []interface{}{"hello"}},

		{Action: "PushObjectKey", Params: []interface{}{"C"}},
		{Action: "BeginObject", Params: []interface{}{}},
		{Action: "PushObjectKey", Params: []interface{}{"foo"}},
		{Action: "PushString", Params: []interface{}{"bar"}},
		{Action: "EndObject", Params: []interface{}{0}},

		{Action: "PushObjectKey", Params: []interface{}{"D"}},
		{Action: "BeginObject", Params: []interface{}{}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushInt", Params: []interface{}{int64(100)}},
		{Action: "EndObject", Params: []interface{}{0}},

		{Action: "PushObjectKey", Params: []interface{}{"E"}},
		{Action: "BeginObject", Params: []interface{}{}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushInt", Params: []interface{}{int64(200)}},
		{Action: "EndObject", Params: []interface{}{0}},

		{Action: "PushObjectKey", Params: []interface{}{"F"}},
		{Action: "BeginArray", Params: []interface{}{}},
		{Action: "PushInt", Params: []interface{}{int64(-1)}},
		{Action: "PushInt", Params: []interface{}{int64(-2)}},
		{Action: "EndArray", Params: []interface{}{0}},

		{Action: "PushObjectKey", Params: []interface{}{"G"}},
		{Action: "BeginArray", Params: []interface{}{}},
		{Action: "BeginObject", Params: []interface{}{}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushInt", Params: []interface{}{int64(300)}},
		{Action: "EndObject", Params: []interface{}{0}},
		{Action: "BeginObject", Params: []interface{}{}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushInt", Params: []interface{}{int64(400)}},
		{Action: "EndObject", Params: []interface{}{0}},
		{Action: "EndArray", Params: []interface{}{0}},

		{Action: "PushObjectKey", Params: []interface{}{"H"}},

		{Action: "BeginArray", Params: []interface{}{}},
		{Action: "BeginObject", Params: []interface{}{}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushInt", Params: []interface{}{int64(500)}},
		{Action: "EndObject", Params: []interface{}{0}},
		{Action: "BeginObject", Params: []interface{}{}},
		{Action: "PushObjectKey", Params: []interface{}{"X"}},
		{Action: "PushInt", Params: []interface{}{int64(600)}},
		{Action: "EndObject", Params: []interface{}{0}},
		{Action: "EndArray", Params: []interface{}{0}},

		{Action: "PushObjectKey", Params: []interface{}{"I"}},
		{Action: "PushInt", Params: []interface{}{int64(456)}},

		{Action: "PushObjectKey", Params: []interface{}{"J"}},
		{Action: "PushInt", Params: []interface{}{int64(1212)}},

		{Action: "EndObject", Params: []interface{}{0}},
	}

	rv := reflect.ValueOf(ow)
	for _, rec := range records {
		act := rv.MethodByName(rec.Action)
		var args []reflect.Value
		for _, p := range rec.Params {
			args = append(args, reflect.ValueOf(p))
		}
		rets := act.Call(args)
		if len(rets) > 0 {
			rErr := rets[len(rets)-1]
			err := rErr.Interface()
			if err != nil {
				t.Fatalf("error at %+v: %+v", rec, err)
			}
		}
	}
	iField := 456
	jField := 1212
	jpField := &jField
	expected := &dummyStruct{
		A: 123,
		B: "hello",
		C: map[string]interface{}{
			"foo": "bar",
		},
		D: innerStruct{
			X: 100,
		},
		E: &innerStruct{
			X: 200,
		},
		F: []int{-1, -2},
		G: []*innerStruct{{X: 300}, {X: 400}},
		H: []innerStruct{{X: 500}, {X: 600}},
		I: &iField,
		J: &jpField,
	}
	if diff := cmp.Diff(target, expected); diff != "" {
		t.Fatal(diff)
	}
}
