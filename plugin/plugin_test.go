// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestScan_ok(t *testing.T) {
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

	tot, err := Scan(tmpDir, ".so")
	if len(tot) != 1 {
		t.Error("unexpected number of plugins found:", tot)
	}
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
}

func TestScan_noFolder(t *testing.T) {
	expectedErr := "open unknown: no such file or directory"
	tot, err := Scan("unknown", "")
	if len(tot) != 0 {
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

func TestScan_emptyFolder(t *testing.T) {
	name, err := ioutil.TempDir(".", "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	tot, err := Scan(name, "")
	if len(tot) != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
	os.RemoveAll(name)
}

func TestScan_noMatches(t *testing.T) {
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
	tot, err := Scan(tmpDir, ".so")
	if len(tot) != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
}
