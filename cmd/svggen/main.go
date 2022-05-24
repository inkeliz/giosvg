package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/inkeliz/giosvg/internal/svgparser"
	"go/format"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var (
	input  string
	output string
	pkg    string
)

func main() {
	flag.StringVar(&input, "i", "", "folder containing svg icons or the path of svg file")
	flag.StringVar(&output, "o", "", "file path to save the go code")
	flag.StringVar(&pkg, "pkg", "", "package name")
	flag.Parse()

	if input == "" {
		panic("invalid input")
	}

	var paths []string
	s, err := os.Stat(input)
	if err != nil {
		panic(err)
	}

	if s.IsDir() {
		if paths, err = filepath.Glob(filepath.Join(input, "*.svg")); err != nil {
			panic(err)
		}
	} else {
		paths = []string{input}
	}

	if pkg == "" {
		if output != "" {
			abs, _ := filepath.Abs(output)
			pkg = filepath.Base(filepath.Dir(abs))
		} else {
			pkg = "assets"
		}
	}

	out := bytes.NewBuffer(nil)
	fmt.Fprintf(out, `package %s

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/inkeliz/giosvg"
	"image/color"
)

var _, _, _, _, _, _ = (*f32.Point)(nil), (*op.Ops)(nil), (*clip.Op)(nil), (*paint.PaintOp)(nil), (*giosvg.Vector)(nil), (*color.NRGBA)(nil)
`, pkg)

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		svg, err := svgparser.ReadIcon(f)
		if err != nil {
			panic(err)
		}

		name := filepath.Base(path)
		name = strings.Replace(name, "-", " ", -1)
		name = strings.Replace(name, "_", " ", -1)
		name = strings.Title(strings.Replace(name, filepath.Ext(name), "", -1))
		name = strings.Replace(name, " ", "", -1)

		fmt.Fprintf(out, `var Vector%s giosvg.Vector = func(ops *op.Ops, w, h float32) {`+"\r\n", name)

		fmt.Fprintf(out, `var (
	size = f32.Point{X: w / %f, Y: h / %f}
	avg = (size.X + size.Y) / 2
	aff = f32.Affine2D{}.Scale(f32.Point{X: float32(0 - %f), Y: float32(0 - %f)}, size)

	end 		clip.PathSpec
	path		clip.Path
	stroke, outline clip.Stack
)`+"\r\n", svg.ViewBox.W, svg.ViewBox.H, svg.ViewBox.X, svg.ViewBox.Y)

		fmt.Fprintf(out, `_, _, _, _, _, _ = avg, aff, end, path, stroke, outline`+"\r\n")

		for _, v := range svg.SVGPaths {
			if v.Style.FillerColor == nil && v.Style.LinerColor == nil {
				continue
			}

			for i, op := range v.Path {
				if i == 0 {
					fmt.Fprintf(out, "\r\n"+`path = clip.Path{}`+"\r\n")
					fmt.Fprintf(out, `path.Begin(ops)`+"\r\n")
				}

				switch op := op.(type) {
				case svgparser.OpMoveTo:
					fmt.Fprintf(out, `path.MoveTo(aff.Transform(f32.Point{X: %f, Y: %f}))`+"\r\n", op.X, op.Y)
				case svgparser.OpLineTo:
					fmt.Fprintf(out, `path.LineTo(aff.Transform(f32.Point{X: %f, Y: %f}))`+"\r\n", op.X, op.Y)
				case svgparser.OpQuadTo:
					fmt.Fprintf(out, `path.QuadTo(aff.Transform(f32.Point{X: %f, Y: %f}), aff.Transform(f32.Point{X: %f, Y: %f}))`+"\r\n", op[0].X, op[0].Y, op[1].X, op[1].Y)
				case svgparser.OpCubicTo:
					fmt.Fprintf(out, `path.CubeTo(aff.Transform(f32.Point{X: %f, Y: %f}), aff.Transform(f32.Point{X: %f, Y: %f}), aff.Transform(f32.Point{X: %f, Y: %f}))`+"\r\n", op[0].X, op[0].Y, op[1].X, op[1].Y, op[2].X, op[2].Y)
				}

				if i == len(v.Path)-1 {
					paint := func(pattern svgparser.Pattern, opacity float64) {
						switch c := pattern.(type) {
						case svgparser.CurrentColor:
							fmt.Fprintf(out, `paint.PaintOp{}.Add(ops)`+"\r\n")
						case svgparser.PlainColor:
							if opacity < 1 {
								c.NRGBA.A = uint8(math.Round(256 * opacity))
							}
							fmt.Fprintf(out, `paint.ColorOp{Color: color.NRGBA{R: %d, G: %d, B: %d, A: %d}}.Add(ops)`+"\r\n", c.NRGBA.R, c.NRGBA.G, c.NRGBA.B, c.NRGBA.A)
							fmt.Fprintf(out, `paint.PaintOp{}.Add(ops)`+"\r\n")
						}
					}

					fmt.Fprintf(out, `end = path.End()`+"\r\n")

					if v.Style.FillerColor != nil {
						fmt.Fprintf(out, `outline = clip.Outline{Path: end}.Op().Push(ops)`+"\r\n")
						paint(v.Style.FillerColor, v.Style.FillOpacity)
						fmt.Fprintf(out, `outline.Pop()`+"\r\n")
					}
					if v.Style.LinerColor != nil {
						fmt.Fprintf(out, `stroke = clip.Stroke{Path: end, Width: %f * avg}.Op().Push(ops)`+"\r\n", v.Style.LineWidth)
						paint(v.Style.LinerColor, v.Style.LineOpacity)
						fmt.Fprintf(out, `stroke.Pop()`+"\r\n")
					}
				}
			}
		}

		fmt.Fprintln(out, `}`)

		f.Close()
	}

	var save io.WriteCloser
	if output == "" {
		save = os.Stdout
	} else {
		if save, err = os.Create(output); err != nil {
			panic(err)
		}
	}
	defer save.Close()

	result, err := format.Source(out.Bytes())
	if err != nil {
		panic(err)
	}

	save.Write(result)
}
