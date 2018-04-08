// +build go1.9

package internal

import "sync"

func NewUntyped() *Untyped {
	return &Untyped{
		data: &sync.Map{},
	}
}

type Untyped struct {
	data *sync.Map
}

func (u *Untyped) Register(name string, v interface{}) {
	u.data.Store(name, v)
}

func (u *Untyped) Get(name string) (interface{}, bool) {
	return u.data.Load(name)
}
