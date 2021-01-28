package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"strings"
)

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
		var visitor astVisitorFunc
		visitor = func(n ast.Node) ast.Visitor {
			switch n := n.(type) {
			case *ast.Field:
				ast.Walk(visitor, n.Type)
				return nil
			case *ast.Ident:
				if _, ok := builtins[n.Name]; ok {
					return nil
				}
				n.Name = pkg + "." + n.Name // bad, but works
				return nil
			case *ast.SelectorExpr:
				// Selectors in type expressions are always qualified names.
				return nil
			}
			return visitor
		}
		ast.Walk(visitor, iface)
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
			}
		}
	}

	if len(iface.Methods.List) == 1 {
		m := iface.Methods.List[0]
		fn, ok := m.Type.(*ast.FuncType)
		if ok {
			fmt.Fprintf(&dst, "type %sFunc %s\n", typ.Name, printGo(fn))
			printMethodDelegate(&dst, &fset, typ.Name.Name+"Func", m.Names[0].Name, fn, true)
			fmt.Fprintln(&dst)
		}
	}

	fmt.Fprintf(&dst, "type %sMock struct {", typ.Name)
	for _, m := range iface.Methods.List {
		_, ok := m.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		fmt.Fprintf(&dst, "\n\t%sFunc ", m.Names[0])
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
		"\nfunc (r %s) %s%s {\n\t",
		recv, method,
		printGo(fn)[len("func"):],
	)
	if fn.Results != nil {
		fmt.Fprint(w, "return ")
	}
	fmt.Fprint(w, "r")
	if !recvIsFunc {
		fmt.Fprintf(w, ".%sFunc", method)
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

type astVisitorFunc func(node ast.Node) (w ast.Visitor)

func (r astVisitorFunc) Visit(node ast.Node) (w ast.Visitor) {
	return r(node)
}
