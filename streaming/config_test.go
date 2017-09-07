package streaming

import (
	"github.com/devopsfaith/krakend/config"
	"testing"
)

func TestStreamConfigGetter_is_present(t *testing.T) {
	getter, ok := config.ConfigGetters[StreamNamespace]
	if !ok {
		t.Error("Nothing stored at the default namespace")
		return
	}
	extra := config.ExtraConfig{
		"Forward": true,
	}
	result := getter(extra)
	res, ok := result.(StreamExtraConfig)
	if !ok {
		t.Error("error casting the returned value")
		return
	}

	if v := res.Forward; !v {
		t.Errorf("unexpected value for key `Forward`: %v", v)
		return
	}
}

func TestStreamConfigGetter_is_not_present(t *testing.T) {
	getter, ok := config.ConfigGetters[StreamNamespace]
	if !ok {
		t.Error("Nothing stored at the default namespace")
		return
	}
	extra := config.ExtraConfig{}
	result := getter(extra)
	res, ok := result.(StreamExtraConfig)
	if !ok {
		t.Error("error casting the returned value")
		return
	}

	if v := res.Forward; v {
		t.Errorf("unexpected value for key `Forward`: %v", v)
		return
	}
}
