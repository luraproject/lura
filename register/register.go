// SPDX-License-Identifier: Apache-2.0

/*
	Package register offers tools for creating and managing registers.
*/
package register

import "sync"

// New returns an initialized Namespaced register
func New() *Namespaced {
	return &Namespaced{data: NewUntyped()}
}

// Namespaced is a register able to keep track of elements stored
// under namespaces and keys
type Namespaced struct {
	data *Untyped
}

// Get returns the Untyped register stored under the namespace
func (n *Namespaced) Get(namespace string) (*Untyped, bool) {
	v, ok := n.data.Get(namespace)
	if !ok {
		return nil, ok
	}
	register, ok := v.(*Untyped)
	return register, ok
}

// Register stores v at the key name of the Untyped register named namespace
func (n *Namespaced) Register(namespace, name string, v interface{}) {
	if register, ok := n.Get(namespace); ok {
		register.Register(name, v)
		return
	}

	register := NewUntyped()
	register.Register(name, v)
	n.data.Register(namespace, register)
}

// AddNamespace adds a new, empty Untyped register under the give namespace (if
// it did not exist)
func (n *Namespaced) AddNamespace(namespace string) {
	if _, ok := n.Get(namespace); ok {
		return
	}
	n.data.Register(namespace, NewUntyped())
}

// NewUntyped returns an empty Untyped register
func NewUntyped() *Untyped {
	return &Untyped{
		data:  map[string]interface{}{},
		mutex: &sync.RWMutex{},
	}
}

// Untyped is a simple register, safe for concurrent access
type Untyped struct {
	data  map[string]interface{}
	mutex *sync.RWMutex
}

// Register stores v under the key name
func (u *Untyped) Register(name string, v interface{}) {
	u.mutex.Lock()
	u.data[name] = v
	u.mutex.Unlock()
}

// Get returns the value stored at the key name
func (u *Untyped) Get(name string) (interface{}, bool) {
	u.mutex.RLock()
	v, ok := u.data[name]
	u.mutex.RUnlock()
	return v, ok
}

// Clone returns a snapshot of the register
func (u *Untyped) Clone() map[string]interface{} {
	u.mutex.RLock()
	res := make(map[string]interface{}, len(u.data))
	for k, v := range u.data {
		res[k] = v
	}
	u.mutex.RUnlock()
	return res
}
