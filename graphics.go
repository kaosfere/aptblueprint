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

// This math is super-duper ugly and was iteratively developed.  I do not
// profess strength in geometry, and there could totally be a better way
// of doing this.   There probably is.   I'm st00pid.  But this works.
func calcPixels(runways []*aptdata.Runway) (pxEndpoints [][2][2]float64, xDimension int, yDimension int) {
	degEndpoints := make([][2]*geo.Point, len(runways))
	pxEndpoints = make([][2][2]float64, len(runways))
	var minLatitude, maxLatitude, minLongitude, maxLongitude,
		xyLengthRatio, xLengthDegrees, yLengthDegrees, lngAdjFactor,
		latAdjFactor float64

	// Set some values we know will be reset by our data to ensure we properly
	// capture all the extremes.  (Setting, say, minLatitude to 0 won't work,
	// because we may have negative latitudes -- instead we'll set it as high
	// as it could possibly be and let our data drag it down.)
	minLatitude = 90
	maxLatitude = -90
	minLongitude = 180
	maxLongitude = -180

	// Find our minimum and maximum lats and longs so we know the
	// dimensions of our bounding box.  Then create a slice of pairs of
	// points representing the ends of each runway.
	for i, r := range runways {
		minLatitude = math.Min(minLatitude,
			math.Min(r.End1Latitude, r.End2Latitude))
		maxLatitude = math.Max(maxLatitude,
			math.Max(r.End1Latitude, r.End2Latitude))
		minLongitude = math.Min(minLongitude,
			math.Min(r.End1Longitude, r.End2Longitude))
		maxLongitude = math.Max(maxLongitude,
			math.Max(r.End1Longitude, r.End2Longitude))

		degEndpoints[i] = [2]*geo.Point{geo.NewPoint(r.End1Latitude, r.End1Longitude),
			geo.NewPoint(r.End2Latitude, r.End2Longitude)}
	}

	// Create a point for each corner of the bounding box
	nwPoint := geo.NewPoint(maxLatitude, minLongitude)
	nePoint := geo.NewPoint(maxLatitude, maxLongitude)
	swPoint := geo.NewPoint(minLatitude, minLongitude)
	sePoint := geo.NewPoint(minLatitude, maxLongitude)

	// Find the lat/long deltas for the X and Y sides.  TODO: Correct this for
	// the slight difference in distance betweenlines like we do below? Meh.
	xLengthDegrees = maxLongitude - minLongitude
	yLengthDegrees = maxLatitude - minLatitude

	// Determine what the length of each side is in feet.  We look for the max
	// here to offset the fact that one degree will have slightly different
	// lengths depending on where in the world you are.  This is probably overly
	// picky.
	xLengthFeet := round(kmToFeet(math.Max(nwPoint.GreatCircleDistance(nePoint),
		swPoint.GreatCircleDistance(sePoint))))
	yLengthFeet := round(kmToFeet(math.Max(nwPoint.GreatCircleDistance(swPoint),
		nePoint.GreatCircleDistance(sePoint))))

	// Determine the ratio of the LengthDegreess of the north/south and
	// east/west sides.  This will be used to scale the actual pixel-counts of
	// the sides.
	xyLengthRatio = float64(xLengthFeet) / float64(yLengthFeet)

	// Set the dimension of each side in pixels witn the longest side being
	// defined by ChartSideLength and the shorter by the xyLengthRatio.
	if xLengthFeet > yLengthFeet {
		xDimension = ChartSideLength
		yDimension = round(ChartSideLength / xyLengthRatio)
	} else {
		yDimension = ChartSideLength
		xDimension = round(ChartSideLength * xyLengthRatio)
	}

	// Eacn axis needs to have its ration of distance in degrees to distance
	// in pixels adjusted seperately, due to the difference in the actual
	// size of a "degree" as mentioned above.  This will create a (faily large)
	// float that we multiply each degree delta by to convert it into pixels.
	lngAdjFactor = float64(xDimension) / xLengthDegrees
	latAdjFactor = float64(yDimension) / yLengthDegrees

	// Now we create a new array, witn the delta degrees converted into pixel
	// coordinates as descibed above.
	for i, r := range degEndpoints {
		pxEndpoints[i] = [2][2]float64{{float64(round((r[0].Lat() - minLatitude) * latAdjFactor)),
			float64(round((r[0].Lng() - minLongitude) * lngAdjFactor))},
			{float64(round((r[1].Lat() - minLatitude) * latAdjFactor)),
				float64(round((r[1].Lng() - minLongitude) * lngAdjFactor))}}
	}

	return pxEndpoints, xDimension, yDimension
}

func drawAirport(runways []*aptdata.Runway, code string, name string, city string, region string, country string) {
	// Get our runway endpoints and image size as pixels.  The math is ugly,
	// so it's contained in a seperate function.
	pxEndpoints, xDimension, yDimension := calcPixels(runways)

	// define the colors for easier setting
	blueprint := color.RGBA{4, 63, 140, 255}
	white := color.RGBA{255, 255, 255, 255}

	// Start with canvas shrink-wrapped to the airport size
	canvas := gg.NewContext(xDimension, yDimension)
	canvas.InvertY()

	// Fill the entire box with blueprint blue
	canvas.SetColor(blueprint)
	canvas.Clear()

	// Prepare for line drawing
	canvas.SetColor(white)
	canvas.SetLineWidth(3)

	// Draw a line for each runway
	for _, p := range pxEndpoints {
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

	// Set strings for our name and location.  Sometimes the city is empty
	// in our data.  Handle that neatly.
	nameCode := fmt.Sprintf("%s (%s)", name, code)
	location := fmt.Sprintf("%s, %s", region, country)
	if city != "" {
		location = fmt.Sprintf("%s, %s", city, location)
	}

	// Get the dimensions of our text
	nWidth, nHeight := canvas.MeasureString(nameCode)
	lWidth, lHeight := canvas.MeasureString(location)

	// And calculate dimensions of its box
	textMargin := float64(10)
	lineSpacing := float64(8)
	textWidth := math.Max(nWidth, lWidth)
	boxWidth := textWidth + float64(textMargin*2)
	boxHeight := nHeight + lHeight + float64(textMargin*2+lineSpacing)

	// Clear out a box for the text to fit into
	boxX := float64(SideLength - boxWidth)
	canvas.DrawRectangle(boxX, 0, boxWidth, boxHeight)
	canvas.SetColor(blueprint)
	canvas.Fill()

	// Add the text
	canvas.SetColor(white)
	canvas.DrawString(nameCode, SideLength-textMargin-nWidth, nHeight+textMargin)
	canvas.DrawString(location, SideLength-textMargin-lWidth, nHeight+lHeight+textMargin+lineSpacing)

	// Finally, draw a border all around it
	canvas.DrawRectangle(0, 0, SideLength, SideLength)
	canvas.Stroke()
	canvas.SavePNG("out.png")
}
