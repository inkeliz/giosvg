package svgdraw

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/inkeliz/giosvg/internal/svgparser"
	"golang.org/x/image/math/fixed"
)

func _f32(a fixed.Point26_6) f32.Point {
	return f32.Pt(float32(a.X.Round()), float32(a.Y.Round()))
}

type Driver struct {
	Op      *op.Ops
	PathOps []*op.Ops
	Index   int
}

func (d *Driver) Reset() {
	d.Index = 0
	d.Op.Reset()
	for _, o := range d.PathOps {
		o.Reset()
	}
}

func (d *Driver) NewPathOp() *op.Ops {
	if len(d.PathOps) <= d.Index {
		d.PathOps = append(d.PathOps, new(op.Ops))
	}
	d.Index++
	return d.PathOps[d.Index-1]
}

func (d *Driver) SetupDrawers(willFill, willStroke bool) (f svgparser.Filler, s svgparser.Stroker) {
	if willFill {
		o := d.NewPathOp()
		path := new(clip.Path)
		path.Begin(o)
		f = &filler{op: d.Op, pathOp: o, path: path}
	}
	if willStroke {
		o := d.NewPathOp()
		path := new(clip.Path)
		path.Begin(o)
		s = &stroker{op: d.Op, pathOp: o, path: path}
	}
	return f, s
}

type filler struct {
	op     *op.Ops
	pathOp *op.Ops
	path   *clip.Path
}

func (f *filler) Clear() {}

func (f *filler) Start(a fixed.Point26_6) {
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
	if closeLoop {
		f.path.Close()
	}
}

func (f *filler) Draw(color svgparser.Pattern, opacity float64) {
	defer op.Save(f.op).Load()

	clip.Outline{Path: f.path.End()}.Op().Add(f.op)
	switch c := color.(type) {
	case svgparser.CurrentColor:
		paint.PaintOp{}.Add(f.op)
	case svgparser.PlainColor:
		paint.Fill(f.op, c.NRGBA)
	}
}

func (f *filler) SetWinding(useNonZeroWinding bool) {}

type stroker struct {
	op      *op.Ops
	path    *clip.Path
	pathOp  *op.Ops
	options clip.StrokeStyle
}

func (s *stroker) Clear() {}

func (s *stroker) Start(a fixed.Point26_6) {
	s.path.Begin(s.pathOp)
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
	if closeLoop {
		s.path.Close()
	}
}

func (s *stroker) Draw(color svgparser.Pattern, opacity float64) {
	defer op.Save(s.op).Load()

	clip.Stroke{Path: s.path.End(), Style: s.options}.Op().Add(s.op)
	switch c := color.(type) {
	case svgparser.CurrentColor:
		paint.PaintOp{}.Add(s.op)
	case svgparser.PlainColor:
		paint.Fill(s.op, c.NRGBA)
	}
}

func (s *stroker) SetStrokeOptions(options svgparser.StrokeOptions) {
	s.options.Width = float32(options.LineWidth.Round())
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
