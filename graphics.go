package main

import "math"
import _ "fmt"
import "github.com/fogleman/gg"

type boundingBox [4]point
type runwayCoords [2]point

func translate(p point, x float64, y float64) point {
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
}

func minLatitude(coords []runwayCoords) float64 {
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
}

func drawAirport(runways []*runway) {
	endpoints := make([]runwayCoords, len(runways))

	for i, r := range runways {
		endpoints[i] = runwayCoords{point{r.End1Latitude, r.End1Longitude},
			point{r.End2Latitude, r.End2Longitude}}
	}

	minRwyLat := minLatitude(endpoints)
	minRwyLong := minLongitude(endpoints)

	for i, r := range endpoints {
		r[0] = round(scale(translate(r[0], minRwyLat, minRwyLong), 10000))
		r[1] = round(scale(translate(r[1], minRwyLat, minRwyLong), 10000))
		endpoints[i] = r
	}

	xLimit := int(maxLongitude(endpoints))
	yLimit := int(maxLatitude(endpoints))

	canvas := gg.NewContext(xLimit, yLimit)
	canvas.InvertY()
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()
	canvas.SetRGB(1, 0, 0)
	canvas.SetLineWidth(1)

	for _, p := range endpoints {
		canvas.DrawLine(p[0].longitude, p[0].latitude,
			p[1].longitude, p[1].latitude)
	}
	canvas.Stroke()
	plot := canvas.Image()

	canvas = gg.NewContext(750, 750)
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()
	canvas.DrawImage(plot, (750-xLimit)/2, (750-yLimit)/2)

	canvas.SavePNG("out.png")

}
