package main

import "fmt"
import _ "math"
import "encoding/csv"
import bolt "github.com/coreos/bbolt"
import "os"
import "io"

const R int = 6371000 //radius of the earth in meters

type point struct {
	latitude  float64
	longitude float64
}

type runway struct {
	name    string
	endA    point
	endB    point
	heading int
	length  int
	width   int
}

type airport struct {
	id      string
	name    string
	runways []runway
}

func initDb() error {
	db, err := bolt.Open("runways.db", 0644, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	runwayFile, err := os.Open("runways.csv")
	if err != nil {
		return err
	}
	defer runwayFile.Close()

	airportFile, err := os.Open("airports.csv")
	if err != nil {
		return err
	}
	defer airportFile.Close()

	airportReader := csv.NewReader(airportFile)
	for {
		record, err := airportReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println(record)
	}

	runwayReader := csv.NewReader(runwayFile)
	for {
		record, err := runwayReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println("RUNWAY:", record)
	}

	return nil

}

func main() {
	err := initDb()
	fmt.Println(err)
	fmt.Println("Done")
}
