package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"git.rcj.io/aptdata"
	"github.com/spf13/viper"
)

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

func doConfig() error {
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetConfigName("aptblueprint")

	viper.SetDefault("datadir", "data")

	return viper.ReadInConfig()
}

func main() {
	err := doConfig()

	if err != nil {
		fmt.Println(err)
		fmt.Println("Error loading config.  Proceeding with defaults.")
	}

	dataDir := viper.GetString("datadir")
	_, err = os.Stat(dataDir)
	if os.IsNotExist(err) {
		fmt.Println("Precreating data directory.")
		err = os.Mkdir(dataDir, 0755)
		if err != nil {
			fmt.Println("Error making data directory:", err)
			os.Exit(1)
		}
	}

	// TODO:  Make this not fail if datadir doesn't already exist
	db, err := aptdata.OpenDB(fmt.Sprintf("%s/%s", dataDir, "aptdata.db"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !db.Populated() {
		fmt.Println("Downloading data.")
		err = aptdata.DownloadData(dataDir)
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
