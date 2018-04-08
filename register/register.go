package register

import (
	"sync"
)

func New() *Namespaced {
	return &Namespaced{
		data:  map[string]*Untyped{},
		mutex: &sync.RWMutex{},
	}
}

type Namespaced struct {
	data  map[string]*Untyped
	mutex *sync.RWMutex
}

func (n *Namespaced) Register(namespace, name string, v interface{}) {
	n.mutex.Lock()
	if r, ok := n.data[namespace]; ok {
		r.Register(name, v)
	} else {
		n.data[namespace] = &Untyped{}
		n.data[namespace].Register(name, v)
	}
	n.mutex.Unlock()
}

func (u *Namespaced) Get(namespace string) (*Untyped, bool) {
	u.mutex.Lock()
	v, ok := u.data[namespace]
	u.mutex.Unlock()
	return v, ok
}

func (n *Namespaced) AddNamespace(namespace string) {
	n.mutex.Lock()
	n.data[namespace] = &Untyped{}
	n.mutex.Unlock()
}

type Untyped struct {
	data sync.Map
}

func (u *Untyped) Register(name string, v interface{}) {
	u.data.Store(name, v)
}

func (u *Untyped) Get(name string) (interface{}, bool) {
	return u.data.Load(name)
}
