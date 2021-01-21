package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestInstamock(t *testing.T) {
	for i, c := range []struct {
		src      string
		expected string
	}{
		{
			src: `
type MyInterface interface {
	foo(ext.Type) (int, MyType)
	bar()
	nene(a, b int, dame Toma, mas ...interface{})
}
`,
			expected: `type MyInterfaceMock struct {
	foo func(ext.Type) (int, mypkg.MyType)
	bar func()
	nene func(a, b int, dame mypkg.Toma, mas ...interface{})
}

func (m MyInterfaceMock) foo(a0 ext.Type) (int, mypkg.MyType) {
	return MyInterfaceMock.foo(a0)
}

func (m MyInterfaceMock) bar() {
	MyInterfaceMock.bar()
}

func (m MyInterfaceMock) nene(a, b int, dame mypkg.Toma, mas ...interface{}) {
	MyInterfaceMock.nene(a, b, dame, mas...)
}
`,
		},
		{
			src: `
type embedder interface {
	MyInterface
	Qux()
}

`,
			expected: `type embedderMock struct {
	Qux func()
}

func (m embedderMock) Qux() {
	embedderMock.Qux()
}
`,
		},
		{
			src: `
type coso interface {
	single(ext.Type) (int, MyType)
}
`,
			expected: `type cosoFunc func(ext.Type) (int, mypkg.MyType)

func (m cosoFunc) single(a0 ext.Type) (int, mypkg.MyType) {
	return cosoFunc(a0)
}
type cosoMock struct {
	single func(a0 ext.Type) (int, mypkg.MyType)
}

func (m cosoMock) single(a0 ext.Type) (int, mypkg.MyType) {
	return cosoMock.single(a0)
}
`,
		},
	} {
		c := c
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()

			got := translate(c.src, "mypkg")
			if c.expected != got {
				t.Errorf("expected %q, got %q", c.expected, got)
			}

			// Now without a package qualifier.

			expected := strings.ReplaceAll(c.expected, "mypkg.", "")
			got = translate(c.src, "")
			if expected != got {
				t.Errorf("expected %q, got %q", expected, got)
			}
		})
	}
}
