// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestRegularLogger(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	bl, _ := NewLogger("DEBUG", buff, "PREFIX")

	bl.Info("A", "B", "C")

	s := buff.String()
	if !strings.Contains(s, "A B C") {
		t.Errorf("missing: A B C : %s", s)
	}
}

func TestNewAnnotatedLogger(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	bl, _ := NewLogger("DEBUG", buff, "PREFIX")

	al, _ := NewAnnotatedLogger(bl, "Mortadelo")
	al.Info("A", "B", "C")

	s := buff.String()
	if !strings.Contains(s, "A B C Mortadelo") {
		t.Errorf("missing suffix: %s", s)
	}
}

func TestWrappedAnnotatedLogger(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	bl, _ := NewLogger("DEBUG", buff, "PREFIX")

	al, _ := NewAnnotatedLogger(bl, "Mortadelo")
	al, _ = NewAnnotatedLogger(al, "Filemon")
	al.Warning("B", "C", "D")

	s := buff.String()
	if !strings.Contains(s, "B C D Mortadelo Filemon") {
		t.Errorf("missing suffix: %s", s)
	}
}
