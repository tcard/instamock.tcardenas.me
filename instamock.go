package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"strings"
	"syscall/js"
)

func main() {
	js.Global().Set("goTranslate", js.FuncOf(func(this js.Value, p []js.Value) (ret interface{}) {
		defer func() {
			r := recover()
			if r != nil {
				ret = fmt.Sprint(r)
			}
		}()
		return translate(p[0].String(), p[1].String())
	}))
	select {}
}

//export translate
func translate(src, pkg string) string {
	var dst strings.Builder

	var fset token.FileSet
	f, err := parser.ParseFile(&fset, "instamock.go", `package instamock; `+src, 0)
	try(err)

	var typ *ast.TypeSpec

declsLoop:
	for _, d := range f.Decls {
		d, ok := d.(*ast.GenDecl)
		if !ok || d.Tok != token.TYPE {
			continue
		}

		for _, spec := range d.Specs {
			typ = spec.(*ast.TypeSpec)
			break declsLoop
		}
	}

	if typ == nil {
		return ""
	}
	iface, ok := typ.Type.(*ast.InterfaceType)
	if !ok {
		return ""
	}

	if pkg != "" {
		// Qualify package names for unqualified, non-builtin named types.
		for _, m := range iface.Methods.List {
			fn, ok := m.Type.(*ast.FuncType)
			if !ok {
				continue
			}
			fields := fn.Params.List
			if fn.Results != nil {
				fields = append(fields, fn.Results.List...)
			}
			for _, field := range fields {
				var id *ast.Ident
				switch t := field.Type.(type) {
				case *ast.Ident:
					id = t
				case *ast.Ellipsis:
					id, _ = t.Elt.(*ast.Ident)
				}
				if id == nil {
					continue
				}
				if _, ok := builtins[id.Name]; ok {
					continue
				}
				id.Name = pkg + "." + id.Name // bad, but works
			}
		}
	}

	if len(iface.Methods.List) == 1 {
		m := iface.Methods.List[0]
		fn, ok := m.Type.(*ast.FuncType)
		if ok {
			fmt.Fprintf(&dst, "type %sFunc %s\n", typ.Name, printGo(fn))
			printMethodDelegate(&dst, &fset, typ.Name.Name+"Func", m.Names[0].Name, fn, true)
		}
	}

	fmt.Fprintf(&dst, "type %sMock struct {", typ.Name)
	for _, m := range iface.Methods.List {
		_, ok := m.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		fmt.Fprintf(&dst, "\n\t%s ", m.Names[0])
		printer.Fprint(&dst, &fset, m.Type)
	}
	fmt.Fprintf(&dst, "\n}\n")

	for _, m := range iface.Methods.List {
		fn, ok := m.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		printMethodDelegate(&dst, &fset, typ.Name.Name+"Mock", m.Names[0].Name, fn, false)
	}

	return dst.String()
}

func printMethodDelegate(
	w io.Writer, fset *token.FileSet,
	recv, method string, fn *ast.FuncType,
	recvIsFunc bool,
) {
	var args []string
	isVariadic := false
	for i, p := range fn.Params.List {
		if len(p.Names) == 0 {
			// Can't pass unnamed arguments to the delegate; make up a name.
			p.Names = append(p.Names, ast.NewIdent(fmt.Sprintf("a%d", i)))
		}
		for _, n := range p.Names {
			args = append(args, n.Name)
		}
		if i == len(fn.Params.List)-1 {
			_, isVariadic = p.Type.(*ast.Ellipsis)
		}
	}
	if isVariadic {
		args[len(args)-1] += "..."
	}

	fmt.Fprintf(w,
		"\nfunc (m %s) %s%s {\n\t",
		recv, method,
		printGo(fn)[len("func"):],
	)
	if fn.Results != nil {
		fmt.Fprint(w, "return ")
	}
	fmt.Fprint(w, recv)
	if !recvIsFunc {
		fmt.Fprintf(w, ".%s", method)
	}

	fmt.Fprintf(w, "(%s)\n}\n", strings.Join(args, ", "))
}

func printGo(node interface{}) string {
	var s strings.Builder
	try(printer.Fprint(&s, new(token.FileSet), node))
	return s.String()
}

func try(err error) {
	if err != nil {
		panic(err)
	}
}

var builtins = map[string]struct{}{
	`bool`:       {},
	`byte`:       {},
	`complex`:    {},
	`complex128`: {},
	`complex64`:  {},
	`error`:      {},
	`float32`:    {},
	`float64`:    {},
	`int`:        {},
	`int16`:      {},
	`int32`:      {},
	`int64`:      {},
	`int8`:       {},
	`iota`:       {},
	`rune`:       {},
	`string`:     {},
	`uint`:       {},
	`uint16`:     {},
	`uint32`:     {},
	`uint64`:     {},
	`uint8`:      {},
	`uintptr`:    {},
}
