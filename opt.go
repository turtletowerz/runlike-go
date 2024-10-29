package main

import (
	"fmt"
	"slices"
	"strings"
)

type option interface {
	Values() []string
}

// Default generic option
type opt[T comparable] struct {
	v    T
	def  T
	name string
}

func (o opt[T]) Values() []string {
	if o.v == o.def {
		return nil
	}

	// If a flag ends wtih "=" or " " then it needs the value. If not it's probably a boolean flag
	if strings.HasSuffix(o.name, "=") || strings.HasSuffix(o.name, " ") {
		return []string{o.name + fmt.Sprintf("%v", o.v)}
	}
	return []string{o.name}
}

// For handling pointers
type optPtr[T comparable] struct {
	v    *T
	def  T
	name string
}

func (o optPtr[T]) Values() []string {
	if o.v == nil {
		return nil
	}
	n := opt[T]{*o.v, o.def, o.name}
	return n.Values()
}

// Slice option for handling slices
type optSlice[T comparable] struct {
	v    []T
	def  []T
	name string
}

// TODO: Some options, like --security-opt, may need to be quoted. Maybe make a separate handler?
func (o optSlice[T]) Values() (ret []string) {
	if o.def == nil {
		o.def = *new([]T)
	}

	if o.v != nil {
		for _, val := range o.v {
			if !slices.Contains(o.def, val) {
				ret = append(ret, o.name+strings.ReplaceAll(fmt.Sprintf("%v", val), "\"", "\\\"")) // TODO
			}
		}
	}
	return
}

type optMap struct {
	v    map[string]string
	name string
}

// TODO: Some options, like --security-opt, may need to be quoted. Maybe make a separate handler?
func (o optMap) Values() (ret []string) {
	for k, v := range o.v {
		ret = append(ret, o.name+k+"="+v)
	}
	return
}

// Allows a custom function to be passed to handle the case.
type optFunc[T any] struct {
	v T
	f func(T) []string
}

func (o optFunc[T]) Values() []string {
	return o.f(o.v)
}
