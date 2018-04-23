package register

import (
	"github.com/devopsfaith/krakend/register/internal"
)

func New() *Namespaced {
	return &Namespaced{NewUntyped()}
}

type Namespaced struct {
	data Untyped
}

func (n *Namespaced) Get(namespace string) (Untyped, bool) {
	v, ok := n.data.Get(namespace)
	if !ok {
		return nil, ok
	}
	register, ok := v.(Untyped)
	return register, ok
}

func (n *Namespaced) Set(namespace, name string, v interface{}) {
	if register, ok := n.Get(namespace); ok {
		register.Set(name, v)
		return
	}

	register := NewUntyped()
	register.Set(name, v)
	n.data.Set(namespace, register)
}

func (n *Namespaced) AddNamespace(namespace string) {
	if _, ok := n.Get(namespace); ok {
		return
	}
	n.data.Set(namespace, NewUntyped())
}

type Untyped interface {
	Set(name string, v interface{})
	Get(name string) (interface{}, bool)
	Clone() map[string]interface{}
}

func NewUntyped() Untyped {
	return internal.NewUntyped()
}
