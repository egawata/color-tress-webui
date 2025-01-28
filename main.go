package main

import (
	"log"
	"syscall/js"
)

type dom struct {
	btnGenerate js.Value
	inputImage  js.Value
}

type app struct {
	el dom
}

func (a *app) Start() {
	a.el.btnGenerate = js.Global().Get("document").Call("getElementById", "generate")
	a.el.btnGenerate.Call("addEventListener", "click", js.FuncOf(a.generate))

	a.el.inputImage = js.Global().Get("document").Call("getElementById", "input-image")
}

func (a *app) generate(this js.Value, p []js.Value) interface{} {
	src := a.el.inputImage.Get("src")
	w := a.el.inputImage.Get("width").Int()
	h := a.el.inputImage.Get("height").Int()
	log.Printf("Input image: %#v (%d x %d)", src, w, h)
	return nil
}

func main() {
	a := app{}
	a.Start()

	<-make(chan struct{})
}
