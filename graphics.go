package main

import "math"
import "fmt"
import "github.com/fogleman/gg"
import "git.rcj.io/aptdata"
import "github.com/kellydunn/golang-geo"
import "image/color"

const SideLength = 640
const OuterMargin = 10
const ChartSideLength = SideLength - OuterMargin

//type boundingBox [4]geo.Point
//type rwyEndpoints [2]geo.Point

/*func translate(p point, x float64, y float64) point {
	p.latitude -= x
	p.longitude -= y
	return p
}

func scale(p point, factor float64) point {
	p.latitude *= factor
	p.longitude *= factor
	return p
}

func round(p point) point {
	p.latitude = math.Floor(p.latitude + 0.5)
	p.longitude = math.Floor(p.longitude + 0.5)
	return p
}*/

/*func minLatitude(coords []runwayCoords) float64 {
	min := coords[0][0].latitude
	for _, c := range coords {
		min = math.Min(min, c[0].latitude)
		min = math.Min(min, c[1].latitude)
	}
	return min
}

func minLongitude(coords []runwayCoords) float64 {
	min := coords[0][0].longitude
	for _, c := range coords {
		min = math.Min(min, c[0].longitude)
		min = math.Min(min, c[1].longitude)
	}
	return min
}

func maxLatitude(coords []runwayCoords) float64 {
	max := coords[0][0].latitude
	for _, c := range coords {
		max = math.Max(max, c[0].latitude)
		max = math.Max(max, c[1].latitude)
	}
	return max
}

func maxLongitude(coords []runwayCoords) float64 {
	max := coords[0][0].longitude
	for _, c := range coords {
		max = math.Max(max, c[0].longitude)
		max = math.Max(max, c[1].longitude)
	}
	return max
}*/

func kmToFeet(km float64) (feet float64) {
	return km * 3280.4
}

func round(raw float64) (rounded int) {
	return int(math.Floor(raw + .5))
}

func drawAirport(runways []*aptdata.Runway) {
	endpoints := make([][2]*geo.Point, len(runways))
	var minLatitude, maxLatitude, minLongitude, maxLongitude float64
	var nePoint, nwPoint, sePoint, swPoint *geo.Point
	// Set some values we know will be reset by our data
	minLatitude = 90
	maxLatitude = -90
	minLongitude = 180
	maxLongitude = -180

	for i, r := range runways {
		minLatitude = math.Min(minLatitude,
			math.Min(r.End1Latitude, r.End2Latitude))
		maxLatitude = math.Max(maxLatitude,
			math.Max(r.End1Latitude, r.End2Latitude))
		minLongitude = math.Min(minLongitude,
			math.Min(r.End1Longitude, r.End2Longitude))
		maxLongitude = math.Max(maxLongitude,
			math.Max(r.End1Longitude, r.End2Longitude))

		endpoints[i] = [2]*geo.Point{geo.NewPoint(r.End1Latitude, r.End1Longitude),
			geo.NewPoint(r.End2Latitude, r.End2Longitude)}
	}

	nwPoint = geo.NewPoint(maxLatitude, minLongitude)
	nePoint = geo.NewPoint(maxLatitude, maxLongitude)
	swPoint = geo.NewPoint(minLatitude, minLongitude)
	sePoint = geo.NewPoint(minLatitude, maxLongitude)

	// TODO: Correct this for the slight difference in distance between
	// lines as well?
	xLongDistance := maxLongitude - minLongitude
	yLatDistance := maxLatitude - minLatitude
	fmt.Println("XLongDistance:", xLongDistance, "YLatDistance:", yLatDistance)

	xDistance := round(kmToFeet(math.Max(nwPoint.GreatCircleDistance(nePoint),
		swPoint.GreatCircleDistance(sePoint))))
	yDistance := round(kmToFeet(math.Max(nwPoint.GreatCircleDistance(swPoint),
		nePoint.GreatCircleDistance(sePoint))))

	var xyDistanceRatio float64
	xyDistanceRatio = float64(xDistance) / float64(yDistance)

	var xDimension, yDimension int

	if xDistance > yDistance {
		xDimension = ChartSideLength
		yDimension = round(ChartSideLength / xyDistanceRatio)
	} else {
		yDimension = ChartSideLength
		xDimension = round(ChartSideLength * xyDistanceRatio)
	}

	fmt.Println("XDistance:", xDistance, "YDistance:", yDistance, "xyDR:", xyDistanceRatio, "xDim:", xDimension, "yDim", yDimension)
	lngAdjFactor := float64(xDimension) / xLongDistance
	latAdjFactor := float64(yDimension) / yLatDistance
	fmt.Println("lngAdj:", lngAdjFactor, "latAdj:", latAdjFactor)

	adjEndpoints := make([][2][2]float64, len(runways))
	for i, r := range endpoints {
		adjEndpoints[i] = [2][2]float64{{float64(round((r[0].Lat() - minLatitude) * latAdjFactor)),
			float64(round((r[0].Lng() - minLongitude) * lngAdjFactor))},
			{float64(round((r[1].Lat() - minLatitude) * latAdjFactor)),
				float64(round((r[1].Lng() - minLongitude) * lngAdjFactor))}}
	}
	fmt.Println(adjEndpoints)

	// DRAWING TIME!

	// define the colors
	blueprint := color.RGBA{4, 63, 140, 255}
	white := color.RGBA{255, 255, 255, 255}

	// Start with canvas shrink-wrapped to the airport size
	canvas := gg.NewContext(xDimension, yDimension)
	canvas.InvertY()

	// Fill the entire box with blueprint blue
	canvas.SetColor(blueprint)
	canvas.Clear()

	canvas.SetColor(white)
	canvas.SetLineWidth(3)

	// Draw a line for each runway
	for _, p := range adjEndpoints {
		canvas.DrawLine(p[0][1], p[0][0],
			p[1][1], p[1][0])
	}

	// Now stroke the lot of them
	canvas.Stroke()

	// Render the chart to an Image
	chart := canvas.Image()

	// Switch to a new context the full size of our picture
	canvas = gg.NewContext(SideLength, SideLength)

	// Fill it with blueprint blue
	canvas.SetColor(blueprint)
	canvas.Clear()

	// Now render the chart image centered in the box
	canvas.DrawImage(chart, (SideLength-xDimension)/2, (SideLength-yDimension)/2)

	// Time to do some labelling!
	canvas.LoadFontFace("flux.ttf", 12)
	canvas.SetLineWidth(3)
	// TODO: Get these from the database
	name := "Chicago Executive Airport (KPWK)"
	location := "Wheeling, IL, USA"

	// Get the dimensions of our text
	nWidth, nHeight := canvas.MeasureString(name)
	lWidth, lHeight := canvas.MeasureString(location)

	// And calculate dimensions of its box
	textMargin := float64(10)
	lineSpacing := float64(8)
	textWidth := math.Max(nWidth, lWidth)
	boxWidth := textWidth + float64(textMargin*2)
	boxHeight := nHeight + lHeight + float64(textMargin*2+lineSpacing)

	// Clear out a box for the text to fit into
	var boxX, boxY float64
	boxX = float64(SideLength - boxWidth)
	boxY = float64(boxHeight)
	canvas.DrawRectangle(boxX, 0, boxWidth, boxHeight)
	canvas.SetColor(blueprint)
	canvas.Fill()

	// And put a border on it
	canvas.DrawLine(boxX, 0, boxX, boxY)
	canvas.DrawLine(boxX, boxY, SideLength, boxY)
	canvas.SetColor(white)
	canvas.Stroke()

	//	canvas.SetRGB(0.016, 0.246, 0.547)
	//canvas.DrawRectangle(0, 0, 200, 200)
	//canvas.Fill()

	canvas.DrawString(name, SideLength-textMargin-nWidth, nHeight+textMargin)
	canvas.DrawString(location, SideLength-textMargin-lWidth, nHeight+lHeight+textMargin+lineSpacing)
	canvas.SavePNG("out.png")
}
