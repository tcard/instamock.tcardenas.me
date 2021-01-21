// +build js

package main

import (
	"fmt"
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
