package process

import (
	"fmt"
	"reflect"
)

func newValueApplyer(v reflect.Value, key *string) (valueApplyer, error) {
	switch v.Kind() {
	case reflect.Struct:
		return ptrApplyer{target: v}, nil
	case reflect.Map:
		return mapApplyer{target: v}, nil
	case reflect.Slice, reflect.Array:
		return sliceApplyer{target: v}, nil
	case reflect.Interface:
		panic("TODO")
	default:
		return nil, fmt.Errorf("unsupported type: %s", v.Type())
	}
}

type valueApplyer interface {
	SetInt(i int64) error
	SetString(str string) error
	BeginArray() (valueApplyer, error)
	EndArray() error
	BeginObject() (valueApplyer, error)
	EndObject() error
	SetKey(k string)
}

type sliceApplyer struct {
	target reflect.Value
}

func (s sliceApplyer) SetInt(i int64) error {
	panic("implement me")
}

func (s sliceApplyer) SetString(str string) error {
	panic("implement me")
}

func (s sliceApplyer) BeginArray() (valueApplyer, error) {
	panic("implement me")
}

func (s sliceApplyer) EndArray() error {
	panic("implement me")
}

func (s sliceApplyer) BeginObject() (valueApplyer, error) {
	panic("implement me")
}

func (s sliceApplyer) EndObject() error {
	panic("implement me")
}

func (s sliceApplyer) SetKey(k string) {
	panic("implement me")
}

type mapApplyer struct {
	target reflect.Value
}

func (m mapApplyer) SetInt(i int64) error {
	panic("implement me")
}

func (m mapApplyer) SetString(str string) error {
	panic("implement me")
}

func (m mapApplyer) BeginArray() (valueApplyer, error) {
	panic("implement me")
}

func (m mapApplyer) EndArray() error {
	panic("implement me")
}

func (m mapApplyer) BeginObject() (valueApplyer, error) {
	panic("implement me")
}

func (m mapApplyer) EndObject() error {
	panic("implement me")
}

func (m mapApplyer) SetKey(k string) {
	panic("implement me")
}

type ptrApplyer struct {
	target reflect.Value
}

func (p ptrApplyer) SetInt(i int64) error {
	panic("implement me")
}

func (p ptrApplyer) SetString(str string) error {
	panic("implement me")
}

func (p ptrApplyer) BeginArray() (valueApplyer, error) {
	panic("implement me")
}

func (p ptrApplyer) EndArray() error {
	panic("implement me")
}

func (p ptrApplyer) BeginObject() (valueApplyer, error) {
	panic("implement me")
}

func (p ptrApplyer) EndObject() error {
	panic("implement me")
}

func (p ptrApplyer) SetKey(k string) {
	panic("implement me")
}
