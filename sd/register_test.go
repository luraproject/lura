// SPDX-License-Identifier: Apache-2.0

package sd

import (
	"testing"

	"github.com/luraproject/lura/v2/config"
)

func TestGetRegister_Register_ok(t *testing.T) {
	sf1 := func(*config.Backend) Subscriber {
		return SubscriberFunc(func() ([]string, error) { return []string{"one"}, nil })
	}
	sf2 := func(*config.Backend) Subscriber {
		return SubscriberFunc(func() ([]string, error) { return []string{"two", "three"}, nil })
	}
	if err := GetRegister().Register("name1", sf1); err != nil {
		t.Error(err)
	}
	if err := GetRegister().Register("name2", sf2); err != nil {
		t.Error(err)
	}

	if h, err := GetRegister().Get("name1")(&config.Backend{SD: "name1"}).Hosts(); err != nil || len(h) != 1 {
		t.Error("error using the sd name1")
	}

	if h, err := GetRegister().Get("name2")(&config.Backend{SD: "name2"}).Hosts(); err != nil || len(h) != 2 {
		t.Error("error using the sd name2")
	}

	if h, err := GetRegister().Get("name2")(&config.Backend{SD: "name2"}).Hosts(); err != nil || len(h) != 2 {
		t.Error("error using the sd name2")
	}

	subscriberFactories = initRegister()
}

func TestGetRegister_Get_unknown(t *testing.T) {
	if h, err := GetRegister().Get("name")(&config.Backend{Host: []string{"name"}}).Hosts(); err != nil || len(h) != 1 {
		t.Error("error using the default sd")
	}
}

func TestGetRegister_Get_errored(t *testing.T) {
	subscriberFactories.data.Register("errored", true)
	if h, err := GetRegister().Get("errored")(&config.Backend{SD: "errored", Host: []string{"name"}}).Hosts(); err != nil || len(h) != 1 {
		t.Error("error using the default sd")
	}
	subscriberFactories = initRegister()
}
