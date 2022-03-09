package main

import (
	_ "embed"
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/inkeliz/giosvg"
)

//go:generate go run github.com/inkeliz/giosvg/cmd/svggen -i "." -o "./school-bus.go" -pkg "main"

func init() {
	//os.Setenv("GIORENDERER", "forcecompute")
}

// Thanks to Freepik from Flaticon Licensed by Creative Commons 3.0 for the example icons shown below.

//go:embed  school-bus.svg
var bus []byte

func main() {

	data := bus

	go func() {
		window := app.NewWindow(app.Title("Gio"))
		defer window.Perform(system.ActionClose)

		vector, err := giosvg.NewVector(data)
		if err != nil {
			panic(err)
		}

		iconRuntime := giosvg.NewIcon(vector)
		iconGenerated := giosvg.NewIcon(VectorSchoolBus)

		ops := new(op.Ops)
		for e := range window.Events() {
			if e, ok := e.(system.FrameEvent); ok {
				gtx := layout.NewContext(ops, e)
				gtx.Constraints.Max.X, gtx.Constraints.Max.Y = 283, 283

				iconRuntime.Layout(gtx)

				offset := op.Offset(layout.FPt(gtx.Constraints.Max)).Push(gtx.Ops)
				iconGenerated.Layout(gtx)
				offset.Pop()

				op.InvalidateOp{}.Add(gtx.Ops)

				e.Frame(ops)
			}
		}
	}()

	app.Main()
}
