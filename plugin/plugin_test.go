package plugin

import (
	"io/ioutil"
	"os"
	"plugin"
	"testing"

	"github.com/devopsfaith/krakend/config"
)

func TestLoad_noFolder(t *testing.T) {
	expectedErr := "open unknown: no such file or directory"
	tot, err := Load(config.Plugin{Folder: "unknown", Pattern: ""})
	if tot != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err == nil {
		t.Error("expecting error!")
		return
	}
	if err.Error() != expectedErr {
		t.Error("unexpected error:", err.Error())
	}
}

func TestLoad_emptyFolder(t *testing.T) {
	name, err := ioutil.TempDir(".", "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	tot, err := Load(config.Plugin{Folder: name, Pattern: ""})
	if tot != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
	os.RemoveAll(name)
}

func TestLoad_noMatches(t *testing.T) {
	tmpDir, err := ioutil.TempDir(".", "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)
	f, err := ioutil.TempFile(tmpDir, "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	f.Close()
	defer os.RemoveAll(tmpDir)
	tot, err := Load(config.Plugin{Folder: tmpDir, Pattern: ".so"})
	if tot != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
}

func TestLoad_erroredLoad(t *testing.T) {
	tmpDir, err := ioutil.TempDir(".", "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)
	f, err := ioutil.TempFile(tmpDir, "test.so")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	f.Close()
	defer os.RemoveAll(tmpDir)
	tot, err := Load(config.Plugin{Folder: tmpDir, Pattern: ".so"})
	if tot != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err == nil {
		t.Error("expecting error!")
		return
	}
	if err.Error()[:55] != "plugin loader found 1 error(s): \nopening plugin 0 (test" {
		t.Error("unexpected error:", err.Error()[:55])
	}
}

func TestLoad_panicRecovered(t *testing.T) {
	intialPOValue := pluginOpener
	pluginOpener = func(path string) (*plugin.Plugin, error) {
		panic("recover this, please")
	}
	defer func() { pluginOpener = intialPOValue }()
	tmpDir, err := ioutil.TempDir(".", "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)
	f, err := ioutil.TempFile(tmpDir, "test.so")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	f.Close()
	defer os.RemoveAll(tmpDir)
	tot, err := Load(config.Plugin{Folder: tmpDir, Pattern: ".so"})
	if tot != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err == nil {
		t.Error("expecting error!")
		return
	}
	if err.Error()[:55] != "plugin loader found 1 error(s): \nopening plugin 0 (test" {
		t.Error("unexpected error:", err.Error()[:55])
	}
}
