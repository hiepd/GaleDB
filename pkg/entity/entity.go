package entity

type Value interface{}

type Key int

type Row struct {
	Key    Key
	Values []Value
}
