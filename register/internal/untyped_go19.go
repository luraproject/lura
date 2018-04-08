// +build go1.9

package internal

import "sync"

type Untyped struct {
	data sync.Map
}

func (u *Untyped) Register(name string, v interface{}) {
	u.data.Store(name, v)
}

func (u *Untyped) Get(name string) (interface{}, bool) {
	return u.data.Load(name)
}
