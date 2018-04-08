// +build !go1.9

package internal

import "sync"

func NewUntyped() *Untyped {
	return &Untyped{
		data:  map[string]interface{}{},
		mutex: &sync.RWMutex{},
	}
}

type Untyped struct {
	data  map[string]interface{}
	mutex *sync.RWMutex
}

func (u *Untyped) Register(name string, v interface{}) {
	u.mutex.Lock()
	u.data[name] = v
	u.mutex.Unlock()
}

func (u *Untyped) Get(name string) (interface{}, bool) {
	u.mutex.RLock()
	v, ok := u.data[name]
	u.mutex.RUnlock()
	return v, ok
}
