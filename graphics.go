package main

import "math"
import "fmt"
import "github.com/fogleman/gg"
import "git.rcj.io/aptdata"
import "github.com/kellydunn/golang-geo"
import "image/color"

const SideLength = 640
const OuterMargin = 100
const ChartSideLength = SideLength - OuterMargin

func kmToFeet(km float64) (feet float64) {
	return km * 3280.4
}

func round(raw float64) (rounded int) {
	return int(math.Floor(raw + .5))
}

func drawAirport(runways []*aptdata.Runway, code string, name string, city string, region string, country string) {
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

	lngAdjFactor := float64(xDimension) / xLongDistance
	latAdjFactor := float64(yDimension) / yLatDistance

	adjEndpoints := make([][2][2]float64, len(runways))
	for i, r := range endpoints {
		adjEndpoints[i] = [2][2]float64{{float64(round((r[0].Lat() - minLatitude) * latAdjFactor)),
			float64(round((r[0].Lng() - minLongitude) * lngAdjFactor))},
			{float64(round((r[1].Lat() - minLatitude) * latAdjFactor)),
				float64(round((r[1].Lng() - minLongitude) * lngAdjFactor))}}
	}

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
	nameCode := fmt.Sprintf("%s (%s)", name, code)
	location := fmt.Sprintf("%s, %s, %s", city, region, country)

	// Get the dimensions of our text
	nWidth, nHeight := canvas.MeasureString(nameCode)
	lWidth, lHeight := canvas.MeasureString(location)

	// And calculate dimensions of its box
	textMargin := float64(10)
	lineSpacing := float64(8)
	textWidth := math.Max(nWidth, lWidth)
	boxWidth := textWidth + float64(textMargin*2)
	boxHeight := nHeight + lHeight + float64(textMargin*2+lineSpacing)
	fmt.Println(boxHeight)

	// Clear out a box for the text to fit into
	//var boxX, boxY float64
	var boxX float64
	boxX = float64(SideLength - boxWidth)
	//boxY = float64(boxHeight)
	canvas.DrawRectangle(boxX, 0, boxWidth, boxHeight)
	canvas.SetColor(blueprint)
	canvas.Fill()

	// And put a border on it
	/*canvas.DrawLine(boxX, 0, boxX, boxY)
	canvas.DrawLine(boxX, boxY, SideLength, boxY)
	canvas.SetColor(white)
	canvas.Stroke()
	*/
	//	canvas.SetRGB(0.016, 0.246, 0.547)
	//canvas.DrawRectangle(0, 0, 200, 200)
	//canvas.Fill()

	canvas.SetColor(white)
	canvas.DrawString(nameCode, SideLength-textMargin-nWidth, nHeight+textMargin)
	canvas.DrawString(location, SideLength-textMargin-lWidth, nHeight+lHeight+textMargin+lineSpacing)

	// Finally, draw a border all around it
	canvas.DrawRectangle(0, 0, SideLength, SideLength)
	canvas.Stroke()
	canvas.SavePNG("out.png")
}
