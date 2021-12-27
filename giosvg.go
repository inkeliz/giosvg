package giosvg

import (
	"bytes"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/inkeliz/giosvg/internal/svgdraw"
	"github.com/inkeliz/giosvg/internal/svgparser"
	"image"
	"io"
)

// IconOp hold the information from the XML/SVG file, in order to avoid
// decoding of the XML.
type IconOp struct {
	render *svgparser.SVGRender
}

// NewIconOpReader creates an IconOp from the given io.Reader. The data is
// expected to be an SVG.
//
// It's safe, and recommended, to reuse the same IconOp across multiples
// Icon. That will avoid the reading of XML/SVG, everytime that the size
// changes.
func NewIconOpReader(reader io.Reader) (*IconOp, error) {
	render, err := svgparser.ReadIcon(reader)
	if err != nil {
		return nil, err
	}

	return &IconOp{render: render}, nil
}

// NewIconOp creates an IconOp from the given data. The data is
// expected to be an SVG.
//
// It's safe, and recommended, to reuse the same IconOp across multiples
// Icon. That will avoid the reading of XML/SVG, everytime that the size
// changes.
func NewIconOp(data []byte) (*IconOp, error) {
	return NewIconOpReader(bytes.NewReader(data))
}

// Icon holds the information of the SVG, and the latest draw.
// It's safe to reuse the same Icon, as long as all of them use the same
// size.
// If the same icon is rendered in different sizes, consider to create
// a new Icon from each one, reusing the same IconOp.
type Icon struct {
	iconOp *IconOp

	driver   *svgdraw.Driver
	lastSize image.Point
	macro    op.CallOp
	op       *op.Ops
}

// NewIcon creates the layout.Widget from the iconOp.
// Similar to widget.List, the Icon keeps the state from the last draw,
// and the drawing is used if the size remains unchanged. You should
// reuse the same Icon across multiples frames.
//
// Make use to not reuse the Icon with different sizes concurrently,
// otherwise the macro will be useless. So, if the same Icon is used twice
// in the same frame, and each one have different sizes, your app will be
// significant slower. You can re-use the same IconOp with multiple Icon.
func NewIcon(iconOp *IconOp) *Icon {
	return &Icon{
		iconOp:   iconOp,
		driver:   &svgdraw.Driver{},
		lastSize: image.Point{},

		// Due to some bug (?) in Gio, that escape the stack,
		// we need to create one macro for each operation.
		//
		// That macro is mandatory to eliminate the allocation
		// by internal/stroke, which is significant.
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
		icon.parserToGio(gtx)
		icon.macro = macro.Stop()
	}

	icon.macro.Add(gtx.Ops)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (icon *Icon) parserToGio(gtx layout.Context) {
	icon.driver.Scale = gtx.Metric.PxPerDp
	icon.driver.Ops = icon.op
	icon.iconOp.render.SetTarget(0, 0, float64(gtx.Constraints.Max.X), float64(gtx.Constraints.Max.Y))
	icon.iconOp.render.Draw(icon.driver, 1.0)
}
