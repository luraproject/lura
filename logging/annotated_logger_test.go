// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewAnnotatedLogger(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	bl, _ := NewLogger("DEBUG", buff, "PREFIX")

	al, _ := NewAnnotatedLogger(bl, "Mortadelo")
	al.Info("A")

	s := buff.String()
	if !strings.HasSuffix(s, "A Mortadelo") {
		t.Errorf("missing suffix: %s", s)
	}
}

func TestWrappedAnnotatedLogger(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	bl, _ := NewLogger("DEBUG", buff, "PREFIX")

	al, _ := NewAnnotatedLogger(bl, "Mortadelo")
	al, _ = NewAnnotatedLogger(al, "Filemon")
	al.Warning("B")

	s := buff.String()
	if !strings.HasSuffix(s, "B Mortadelo Filemon") {
		t.Errorf("missing suffix: %s", s)
	}
}
