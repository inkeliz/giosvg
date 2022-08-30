package giosvg

import (
	"bytes"
	"image"
	"io"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/inkeliz/giosvg/internal/svgdraw"
	"github.com/inkeliz/giosvg/internal/svgparser"
)

// Vector hold the information from the XML/SVG file, in order to avoid
// decoding of the XML.
type Vector func(ops *op.Ops, w, h float32)

// Layout implements layout.Widget, that renders the current vector without any cache.
// Consider using NewIcon instead.
//
// You should avoid it, that functions only exists to simplify integration
// to custom cache implementations.
func (v Vector) Layout(gtx layout.Context) layout.Dimensions {
	v(gtx.Ops, float32(gtx.Constraints.Max.X), float32(gtx.Constraints.Max.Y))
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

// NewVector creates an IconOp from the given data. The data is
// expected to be an SVG/XML
func NewVector(data []byte) (Vector, error) {
	return NewVectorReader(bytes.NewReader(data))
}

// NewVectorReader creates an IconOp from the given io.Reader. The data is
// expected to be an SVG/XML
func NewVectorReader(reader io.Reader) (Vector, error) {
	render, err := svgparser.ReadIcon(reader)
	if err != nil {
		return nil, err
	}

	return func(ops *op.Ops, w, h float32) {
		render.SetTarget(0-render.ViewBox.X, 0-render.ViewBox.Y, float64(w), float64(h))
		scale := (float32(float64(w)/render.ViewBox.W) + float32(float64(h)/render.ViewBox.H)) / 2
		render.Draw(&svgdraw.Driver{Ops: ops, Scale: scale}, 1.0)
	}, nil
}

// Icon keeps a cache from the last frame and re-uses it if
// the size didn't change.
type Icon struct {
	vector Vector

	lastSize image.Point
	macro    op.CallOp
	op       *op.Ops
}

// NewIcon creates the layout.Widget from the iconOp.
// Similar to widget.List, the Icon keeps the state from the last draw,
// and the drawing is used if the size remains unchanged. You should
// reuse the same Icon across multiples frames.
//
// Make sure to not reuse the Icon with different sizes in the same frame,
// if the same Icon is used twice  in the same frame you MUST create
// two Icon, for each one.
func NewIcon(vector Vector) *Icon {
	return &Icon{
		vector:   vector,
		lastSize: image.Point{},

		op: new(op.Ops),
	}
}

// Layout implements widget.Layout.
// It will render the icon based on the given layout.Constraints.Max.
// If the SVG uses `currentColor` you can set the color using
// paint.ColorOp.
func (icon *Icon) Layout(gtx layout.Context) layout.Dimensions {
	if icon.lastSize != gtx.Constraints.Max {
		// If the size changes, we can't re-use the macro.
		icon.lastSize = gtx.Constraints.Max

		icon.op.Reset()
		macro := op.Record(icon.op)
		icon.vector(icon.op, float32(gtx.Constraints.Max.X), float32(gtx.Constraints.Max.Y))
		icon.macro = macro.Stop()
	}

	icon.macro.Add(gtx.Ops)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}
