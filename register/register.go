package register

import (
	"sync"

	"github.com/devopsfaith/krakend/register/internal"
)

func New() *Namespaced {
	return &Namespaced{
		data:  map[string]*internal.Untyped{},
		mutex: &sync.RWMutex{},
	}
}

type Namespaced struct {
	data  map[string]*internal.Untyped
	mutex *sync.RWMutex
}

func (u *Namespaced) Get(namespace string) (*internal.Untyped, bool) {
	u.mutex.Lock()
	v, ok := u.data[namespace]
	u.mutex.Unlock()
	return v, ok
}

func (n *Namespaced) Register(namespace, name string, v interface{}) {
	n.mutex.Lock()
	if r, ok := n.data[namespace]; ok {
		r.Register(name, v)
	} else {
		n.data[namespace] = new(internal.Untyped)
		n.data[namespace].Register(name, v)
	}
	n.mutex.Unlock()
}

func (n *Namespaced) AddNamespace(namespace string) {
	n.mutex.Lock()
	n.data[namespace] = new(internal.Untyped)
	n.mutex.Unlock()
}
