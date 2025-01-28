package main

import (
	"fmt"
	"syscall/js"
)

type dom struct {
	btnGenerate js.Value
}

type app struct {
	el dom
}

func (a *app) Start() {
	a.el.btnGenerate = js.Global().Get("document").Call("getElementById", "generate")
	a.el.btnGenerate.Call("addEventListener", "click", js.FuncOf(a.generate))
}

func (a *app) generate(this js.Value, p []js.Value) interface{} {
	fmt.Println("Generate Button clicked")
	return nil
}

func main() {
	a := app{}
	a.Start()

	<-make(chan struct{})
}
