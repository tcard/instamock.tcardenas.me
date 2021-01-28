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
			expected: `
				type MyInterfaceMock struct {
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
			expected: `
				type embedderMock struct {
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
			expected: `
				type cosoFunc func(ext.Type) (int, mypkg.MyType)

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

		{
			src: `
				type deepNames interface {
					foo(*MyType, []MyType, []*[123]MyType, not.MyType)
					bar(arg struct{
						f func([]MyType) []MyType
					})
				}
			`,
			expected: `
				type deepNamesMock struct {
					fooFunc func(*mypkg.MyType, []mypkg.MyType, []*[123]mypkg.MyType, not.MyType)
					barFunc func(arg struct {
					f func([]mypkg.MyType) []mypkg.MyType
				})
				}

				func (r deepNamesMock) foo(a0 *mypkg.MyType, a1 []mypkg.MyType, a2 []*[123]mypkg.MyType, a3 not.MyType) {
					r.fooFunc(a0, a1, a2, a3)
				}

				func (r deepNamesMock) bar(arg struct {	f func([]mypkg.MyType) []mypkg.MyType}) {
					r.barFunc(arg)
				}
			`,
		},
	} {
		c := c
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()

			equal := func(expected, got string) {
				t.Helper()
				replace := strings.NewReplacer(
					"\n", "",
					"\t", "",
					" ", "",
				).Replace
				if replace(expected) != replace(got) {
					t.Errorf("expected:\n\n%s\ngot:\n\n%s\n%[1]q\n%[2]q", replace(expected), replace(got))
				}
			}

			got := translate(c.src, "mypkg")
			equal(c.expected, got)

			// Now without a package qualifier.

			expected := strings.ReplaceAll(c.expected, "mypkg.", "")
			got = translate(c.src, "")
			equal(expected, got)
		})
	}
}
