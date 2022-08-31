package main

import (
	_ "embed"
	"image"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/inkeliz/giosvg"
)

//go:generate go run github.com/inkeliz/giosvg/cmd/svggen -i "." -o "./school-bus.go" -pkg "main"

func init() {
	// os.Setenv("GIORENDERER", "forcecompute")
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
				gtx.Constraints.Max.X = gtx.Constraints.Max.X / 2
				gtx.Constraints.Min = gtx.Constraints.Max

				layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min = image.Point{} // Keep aspect ratio.
					return iconRuntime.Layout(gtx)
				})

				offset := op.Offset(f32.Pt(float32(gtx.Constraints.Max.X), 0)).Push(gtx.Ops)
				layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min = image.Point{} // Keep aspect ratio.
					return iconGenerated.Layout(gtx)
				})
				offset.Pop()

				op.InvalidateOp{}.Add(gtx.Ops)

				e.Frame(ops)
			}
		}
	}()

	app.Main()
}
