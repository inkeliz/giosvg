package svgdraw

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/inkeliz/giosvg/internal/svgparser"
	"golang.org/x/image/math/fixed"
	"math"
)

func _f32(a fixed.Point26_6) f32.Point {
	return f32.Pt(float32(a.X.Round()), float32(a.Y.Round()))
}

type Driver struct {
	Ops   []*DrawOp
	Scale float32
	index int
}

func (d *Driver) Reset() {
	for i := range d.Ops {
		d.Ops[i].Ops.Reset()
		d.Ops[i].Color = nil
	}
	d.index = 0
}

type DrawOp struct {
	*op.Ops
	Path  clip.Op
	Color *paint.ColorOp
}

func (d *Driver) NewDrawOp() *DrawOp {
	defer func() { d.index++ }()
	if len(d.Ops) > d.index {
		return d.Ops[d.index]
	}

	o := &DrawOp{Ops: new(op.Ops)}
	d.Ops = append(d.Ops, o)
	return o
}

func (d *Driver) SetupDrawers(willFill, willStroke bool) (f svgparser.Filler, s svgparser.Stroker) {
	if willFill {
		f = &filler{pathOp: d.NewDrawOp()}
	}
	if willStroke {
		s = &stroker{pathOp: d.NewDrawOp(), scale: d.Scale}
	}
	return f, s
}

type filler struct {
	path   *clip.Path
	pathOp *DrawOp
}

func (f *filler) Clear() {}

func (f *filler) Start(a fixed.Point26_6) {
	if f.path == nil {
		f.path = new(clip.Path)
		f.path.Begin(f.pathOp.Ops)
	}

	f.path.MoveTo(_f32(a))
}

func (f *filler) Line(b fixed.Point26_6) {
	f.path.LineTo(_f32(b))
}

func (f *filler) QuadBezier(b, c fixed.Point26_6) {
	f.path.QuadTo(_f32(b), _f32(c))
}

func (f *filler) CubeBezier(b, c, d fixed.Point26_6) {
	f.path.CubeTo(_f32(b), _f32(c), _f32(d))
}

func (f *filler) Stop(closeLoop bool) {
	if f.path != nil {
		f.path.Close()
	}
}

func (f *filler) Draw(color svgparser.Pattern, opacity float64) {
	f.pathOp.Path = clip.Outline{Path: f.path.End()}.Op()

	defer op.Save(f.pathOp.Ops).Load()
	switch c := color.(type) {
	case svgparser.CurrentColor:
		// NO-OP
	case svgparser.PlainColor:
		if opacity < 1 {
			c.NRGBA.A = uint8(math.Round(256 * opacity))
		}
		f.pathOp.Color = &paint.ColorOp{Color: c.NRGBA}
	}
}

func (f *filler) SetWinding(useNonZeroWinding bool) {}

type stroker struct {
	path    *clip.Path
	options clip.StrokeStyle
	pathOp  *DrawOp
	scale   float32
}

func (s *stroker) Clear() {}

func (s *stroker) Start(a fixed.Point26_6) {
	if s.path == nil {
		s.path = new(clip.Path)
		s.path.Begin(s.pathOp.Ops)
	}

	s.path.MoveTo(_f32(a))
}

func (s *stroker) Line(b fixed.Point26_6) {
	s.path.LineTo(_f32(b))
}

func (s *stroker) QuadBezier(b, c fixed.Point26_6) {
	s.path.QuadTo(_f32(b), _f32(c))
}

func (s *stroker) CubeBezier(b, c, d fixed.Point26_6) {
	s.path.CubeTo(_f32(b), _f32(c), _f32(d))
}

func (s *stroker) Stop(closeLoop bool) {
	if s.path != nil && closeLoop {
		s.path.Close()
	}
}

func (s *stroker) Draw(color svgparser.Pattern, opacity float64) {
	s.pathOp.Path = clip.Stroke{Path: s.path.End(), Style: s.options}.Op()

	defer op.Save(s.pathOp.Ops).Load()
	switch c := color.(type) {
	case svgparser.CurrentColor:
		// NO-OP
	case svgparser.PlainColor:
		if opacity < 1 {
			c.NRGBA.A = uint8(math.Round(256 * opacity))
		}
		s.pathOp.Color = &paint.ColorOp{Color: c.NRGBA}
	}
}

func (s *stroker) SetStrokeOptions(options svgparser.StrokeOptions) {
	s.options.Width = float32(options.LineWidth.Round()) * s.scale
	s.options.Miter = float32(options.Join.MiterLimit.Round())

	switch options.Join.LeadLineCap {
	case svgparser.SquareCap:
		s.options.Cap = clip.SquareCap
	case svgparser.RoundCap:
		s.options.Cap = clip.RoundCap
	default:
		s.options.Cap = clip.FlatCap
	}

	switch options.Join.LineJoin {
	case svgparser.Bevel:
		s.options.Join = clip.BevelJoin
	default:
		s.options.Join = clip.RoundJoin
	}
}
