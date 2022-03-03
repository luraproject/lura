// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"github.com/luraproject/lura/v2/register"
)

func NewRegister() *Register {
	return &Register{
		responseCombiners,
	}
}

type Register struct {
	*combinerRegister
}

type combinerRegister struct {
	data     *register.Untyped
	fallback ResponseCombiner
}

func newCombinerRegister(data map[string]ResponseCombiner, fallback ResponseCombiner) *combinerRegister {
	r := register.NewUntyped()
	for k, v := range data {
		r.Register(k, v)
	}
	return &combinerRegister{r, fallback}
}

func (r *combinerRegister) GetResponseCombiner(name string) (ResponseCombiner, bool) {
	v, ok := r.data.Get(name)
	if !ok {
		return r.fallback, ok
	}
	if rc, ok := v.(ResponseCombiner); ok {
		return rc, ok
	}
	return r.fallback, ok
}

func (r *combinerRegister) SetResponseCombiner(name string, rc ResponseCombiner) {
	r.data.Register(name, rc)
}
