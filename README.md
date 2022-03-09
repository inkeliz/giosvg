GIOSVG
-----

Give to your app some SVG icons. (:

------------

Example:

    // Embed your SVG into Golang:
    //go:embed your_icon.svg
    var iconFile []byte

	// Give your SVG/XML to the Vector:
	vector, err := giosvg.NewVector(iconFile)
	if err != nil {
		panic(err)
	}
	
	// Create the Icon:
	icon := giosvg.NewIcon(vector)

    func someWidget(gtx layout.Context) layout.Dimensions {
	    // Render your icon anywhere:
        return icon.Layout(gtx)
    }

You can use `embed` to include your icon. The `Vector` can be reused to avoid parse the SVG multiple times.

If your icon uses `currentColor`, you can use `paint.ColorOp`:

    func someWidget(gtx layout.Context) layout.Dimensions {
	    // Render your icon anywhere, with custom color:
        paint.ColorOp{Color: color.NRGBA{B: 255, A: 255}}.Add(gtx.Ops)
        return icon.Layout(gtx)
    }


------------

The icons in the `example` are from Freepik from Flaticon Licensed by Creative Commons 3.0.

----------

That work is based on OKSVG.
