package main

import _ "github.com/pkg/errors"
import "fmt"
import _ "github.com/coreos/bbolt"
import _ "github.com/vmihailenco/msgpack"
import "git.rcj.io/aptdata"
import "os"

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
	runways, err := db.GetRunways("KPWK")
	runways = filterForCoords(runways)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, x := range runways {
		fmt.Println(x)
	}

	drawAirport(runways)

}
