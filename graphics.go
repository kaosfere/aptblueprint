package main

import "math"
import "fmt"
import "github.com/fogleman/gg"
import "git.rcj.io/aptdata"
import "github.com/kellydunn/golang-geo"

const LongSide = 1000

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
		xDimension = LongSide
		yDimension = round(LongSide / xyDistanceRatio)
	} else {
		yDimension = LongSide
		xDimension = round(LongSide * xyDistanceRatio)
	}

	fmt.Println("XDistance:", xDistance, "YDistance:", yDistance, "xyDR:", xyDistanceRatio, "xDim:", xDimension, "yDim", yDimension)

	/*
		fmt.Println("xDistanceFeet:", xDistance, "yDistanceFeet:", yDistance)
		xScaleFactor := xDistance / XSize
		yScaleFactor := yDistance / YSize
		scaleFactor := math.Max(xScaleFactor, yScaleFactor)
		XYRatio := xDistance / yDistance
		fmt.Println("XYRatio:", XYRatio)
		XYLLRatio := xLongDistance / yLatDistance
		fmt.Println("XYLLRatio:", XYLLRatio)

		fmt.Println(xScaleFactor, yScaleFactor, "max:", scaleFactor)

		adjEndpoints := make([][2][2]float64, len(runways))
		yAdjFactor := 1000 / (math.Max(xLongDistance, yLatDistance))
		xAdjFactor := yAdjFactor * XYRatio
		fmt.Println("xadjfactor:", xAdjFactor, "yadjfactor", yAdjFactor)

		for i, r := range endpoints {
			adjEndpoints[i] = [2][2]float64{{round((r[0].Lat() - minLatitude) * xAdjFactor),
				round((r[0].Lng() - minLongitude) * yAdjFactor)},
				{round((r[1].Lat() - minLatitude) * xAdjFactor),
					round((r[1].Lng() - minLongitude) * yAdjFactor)}}
		}
		fmt.Println(adjEndpoints)
	*/
	canvas := gg.NewContext(xDimension, yDimension)
	canvas.InvertY()
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()
	canvas.SetRGB(1, 0, 0)
	canvas.SetLineWidth(1)
	/*
		for _, p := range adjEndpoints {
			canvas.DrawLine(p[0][1], p[0][0],
				p[1][1], p[1][0])
		}
		canvas.Stroke()
		//	plot := canvas.Image()
		//	fmt.Println("BOUNDS ARE", plot.Bounds())

		//	canvas = gg.NewContext(1200, 1200)
		//	canvas.SetRGB(1, 1, 1)
		//	canvas.Clear()
		//	canvas.DrawImage(plot, (1200-XSize)/2, (1200-YSize)/2)
	*/
	canvas.SavePNG("out.png")
}
