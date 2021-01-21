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
	fooFunc func(ext.Type) (int, mypkg.MyType)
	barFunc func()
	neneFunc func(a, b int, dame mypkg.Toma, mas ...interface{})
}

func (r MyInterfaceMock) foo(a0 ext.Type) (int, mypkg.MyType) {
	return r.fooFunc(a0)
}

func (r MyInterfaceMock) bar() {
	r.barFunc()
}

func (r MyInterfaceMock) nene(a, b int, dame mypkg.Toma, mas ...interface{}) {
	r.neneFunc(a, b, dame, mas...)
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
	QuxFunc func()
}

func (r embedderMock) Qux() {
	r.QuxFunc()
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

func (r cosoFunc) single(a0 ext.Type) (int, mypkg.MyType) {
	return r(a0)
}
type cosoMock struct {
	singleFunc func(a0 ext.Type) (int, mypkg.MyType)
}

func (r cosoMock) single(a0 ext.Type) (int, mypkg.MyType) {
	return r.singleFunc(a0)
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
