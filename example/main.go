package main

import (
	_ "embed"
	"fmt"
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/inkeliz/giosvg"
	"image/color"
	"math/rand"
	"time"
)

func init() {
	//os.Setenv("GIORENDERER", "forcecompute")
}

// Thanks to Freepik from Flaticon Licensed by Creative Commons 3.0 for the example icons shown below.

//go:embed  school-bus.svg
var bus []byte

func main() {

	data := bus

	go func() {
		window := app.NewWindow(
			app.Title("Gio"),
			app.Size(unit.Dp(393), unit.Dp(851)),
			app.MinSize(unit.Dp(393), unit.Dp(351)),
			app.NavigationColor(color.NRGBA{R: 255, G: 8, B: 90, A: 255}),
			app.StatusColor(color.NRGBA{R: 255, G: 8, B: 90, A: 255}),
		)
		defer window.Close()


		t := time.Now()
		render, err := giosvg.NewIconOp(data)
		if err != nil {
			panic(err)
		}
		fmt.Println("parse time", time.Since(t))


		icon := giosvg.NewIcon(render)
		ops := new(op.Ops)
		for e := range window.Events() {
			if e, ok := e.(system.FrameEvent); ok {
				gtx := layout.NewContext(ops, e)

				gtx.Constraints.Max.X, gtx.Constraints.Max.Y = rand.Intn(gtx.Constraints.Max.X), rand.Intn(gtx.Constraints.Max.Y)
				icon.Layout(gtx)

				op.InvalidateOp{At: time.Now().Add(1 * time.Second)}.Add(gtx.Ops)

				e.Frame(ops)
			}
		}
	}()

	app.Main()
}
