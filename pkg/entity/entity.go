package entity

import "reflect"

type Value interface{}

type Key int

type Row struct {
	Key    Key
	Values []Value
}

type Column struct {
	Kind reflect.Kind
	Name string
}
