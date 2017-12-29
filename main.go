package main

import "fmt"
import "git.rcj.io/aptdata"
import "os"
import "math/rand"
import "time"

type point struct {
	latitude  float64
	longitude float64
}

func filterForCoords(raw []*aptdata.Runway) []*aptdata.Runway {
	filtered := []*aptdata.Runway{}
	for _, r := range raw {
		if (r.End1Latitude == 0 && r.End1Longitude == 0) ||
			(r.End2Latitude == 0 && r.End2Longitude == 0) {
			continue
		}
		filtered = append(filtered, r)
	}

	return filtered
}

func main() {
	db, err := aptdata.OpenDB("aptdata.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !db.Populated() {
		fmt.Println("Loading DB")
		err = db.Load("data")
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var code string
	var runways []*aptdata.Runway
	minRunways := 2

	codes, err := db.GetCodes()
	rand.Seed(time.Now().Unix())
	for {
		code = codes[rand.Intn(len(codes))]
		runways, err = db.GetRunways(code)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		runways = filterForCoords(runways)
		fmt.Println(len(runways), "runways found")
		if len(runways) >= minRunways {
			break
		}
	}

	runways = filterForCoords(runways)

	apt, err := db.GetAirport(code)
	name := apt.Name
	city := apt.City

	region, err := db.GetRegion(apt.Region)
	country, err := db.GetCountry(apt.Country)

	drawAirport(runways, code, name, city, region.Name, country.Name)

}
