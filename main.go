package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/kaosfere/aptdata"
	"github.com/spf13/viper"
)

type point struct {
	latitude  float64
	longitude float64
}

func randomAirport(db *aptdata.AptDB) (code string, err error) {
	var runways []*aptdata.Runway
	minRunways := 2

	codes, err := db.GetCodes()
	rand.Seed(time.Now().Unix())
	for {
		code = codes[rand.Intn(len(codes))]
		runways, err = db.GetRunways(code)

		if err != nil {
			return code, err
		}

		runways = filterForCoords(runways)
		if len(runways) >= minRunways {
			break
		}
	}

	return code, nil
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
	viper.SetDefault("outdir", ".")
	viper.SetDefault("font", "flux.ttf")

	return viper.ReadInConfig()
}

func doDownload() error {
	err := aptdata.DownloadData(viper.GetString("datadir"))
	return err
}

func doLoad() error {
	dataDir := viper.GetString("datadir")
	db, err := aptdata.OpenDB(fmt.Sprintf("%s/%s", dataDir, "aptdata.db"))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Reload(dataDir)
}

func doGenerate(ident string) (*aptdata.Airport, error) {
	dataDir := viper.GetString("datadir")
	db, err := aptdata.OpenDB(fmt.Sprintf("%s/%s", dataDir, "aptdata.db"))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if !db.Populated() {
		return nil, fmt.Errorf("database not populated")
	}

	if ident == "" {
		ident, err = randomAirport(db)
		if err != nil {
			return nil, fmt.Errorf("Error picking airport: %s", err)
		}
	}

	apt, err := db.GetAirport(ident)
	if err != nil {
		return apt, err
	}

	runways, err := db.GetRunways(ident)
	if err != nil {
		return apt, err
	}

	name := apt.Name
	city := apt.City

	region, err := db.GetRegion(apt.Region)
	if err != nil {
		return apt, err
	}

	country, err := db.GetCountry(apt.Country)
	if err != nil {
		return apt, err
	}

	drawAirport(runways, ident, name, city, region.Name, country.Name)
	return apt, nil
}

func main() {
	err := doConfig()
	var apt *aptdata.Airport

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

	if len(os.Args) == 1 {
		fmt.Println("Generating random airport.")
		_, err = doGenerate("")
		if err != nil {
			fmt.Println("Error generating airport: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// TODO: The interface here between genrate/post and the database feels
	// a bit awkward, and could stand to be refactored.  If nothing else,
	// it lacks a little bit of DRY.
	switch os.Args[1] {
	case "download":
		fmt.Println("Downloading data.")
		err = doDownload()
	case "load", "reload":
		fmt.Println("Loading database.")
		err = doLoad()
	case "generate":
		if len(os.Args) > 2 {
			ident := os.Args[2]
			fmt.Printf("Generating %s.\n", ident)
			apt, err = doGenerate(ident)
		} else {
			fmt.Println("Generating random airport.")
			apt, err = doGenerate("")
		}
	case "post":
		if len(os.Args) > 2 {
			ident := os.Args[2]
			fmt.Printf("Generating %s.\n", ident)
			apt, err = doGenerate(ident)
		} else {
			fmt.Println("Generating random airport.")
			apt, err = doGenerate("")
		}

		creds := credentials{viper.GetString("consumer_key"),
			viper.GetString("consumer_secret"), viper.GetString("access_token"),
			viper.GetString("access_token_secret")}
		err = post(creds, apt)
	default:
		fmt.Printf("%s [download|reload|generate]\n", os.Args[0])
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
