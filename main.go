package main

import "github.com/pkg/errors"
import "fmt"
import "github.com/coreos/bbolt"
import "github.com/vmihailenco/msgpack"
import "os"
import "math"

type point struct {
	latitude  float64
	longitude float64
}

func buildDB(rebuild bool) (*bolt.DB, error) {
	db, err := bolt.Open("aptdata.db", 0644, nil)
	if err != nil {
		return nil, errors.Wrap(err, "database open failed")
	}
	//defer db.Close()

	if rebuild {
		err = db.Update(func(tx *bolt.Tx) error {
			err := tx.DeleteBucket([]byte("Airports"))
			if err != nil {
				if err.Error() != "bucket not found" {
					return errors.Wrap(err, "airports bucket")
				}
			}
			err = tx.DeleteBucket([]byte("Runways"))
			if err != nil {
				if err.Error() != "bucket not found" {
					return errors.Wrap(err, "runways bucket")
				}
			}
			return nil
		})

		if err != nil {
			return nil, errors.Wrap(err, "database cleanup")
		}

		err = loadAirports(db)
		if err != nil {
			return nil, errors.Wrap(err, "loading airports")
		}
		err = loadRunways(db)
		if err != nil {
			return nil, errors.Wrap(err, "loading runways")
		}

	}
	return db, err
}

func getRunways(db *bolt.DB, ident string) ([]*runway, error) {
	var runways []*runway
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Runways"))
		b2 := b.Bucket([]byte(ident))
		b2.ForEach(func(k, v []byte) error {
			var rwy runway
			msgpack.Unmarshal(v, &rwy)
			runways = append(runways, &rwy)
			return nil
		})
		return nil
	})

	if err != nil {
		return runways, errors.Wrap(err, "get runways")
	}

	return runways, nil
}

func runwayBoundingBox(runways []*runway) [4]point {
	var maxLatitude, minLatitude, maxLongitude, minLongitude float64
	minLatitude = 90
	minLongitude = 180
	maxLatitude = -90
	maxLongitude = -180

	for _, runway := range runways {
		rwyMaxLatitude := math.Max(runway.End1Latitude, runway.End2Latitude)
		rwyMinLatitude := math.Min(runway.End1Latitude, runway.End2Latitude)
		rwyMaxLongitude := math.Max(runway.End1Longitude, runway.End2Longitude)
		rwyMinLongitude := math.Min(runway.End1Longitude, runway.End2Longitude)

		maxLatitude = math.Max(maxLatitude, rwyMaxLatitude)
		minLatitude = math.Min(minLatitude, rwyMinLatitude)
		maxLongitude = math.Max(maxLongitude, rwyMaxLongitude)
		minLongitude = math.Min(minLongitude, rwyMinLongitude)
	}

	return [4]point{
		point{maxLatitude, minLongitude},
		point{maxLatitude, maxLongitude},
		point{minLatitude, minLongitude},
		point{minLatitude, maxLongitude}}
}

func main() {
	db, err := buildDB(false)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	/*	err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Airports"))
			aptg := b.Get([]byte("KORD"))
			if aptg == nil {
				fmt.Println("NADA")
				return nil
			}
			msgpack.Unmarshal(aptg, &apt)
			fmt.Println(apt)

			b = tx.Bucket([]byte("Runways"))
			wayg := b.Bucket([]byte("KORD"))
			wayg.ForEach(func(k, v []byte) error {
				fmt.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
			if wayg == nil {
				fmt.Println("NADA BUCKET")
				return nil
			}

				fmt.Println("sub bucket", wayg)
				msgpack.Unmarshal(wayg, &rwy)
				fmt.Println(rwy)
			return nil
		})
		if err != nil {
			fmt.Println("POOP", err)
		}
	*/
	runways, err := getRunways(db, "KORD")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, runway := range runways {
		fmt.Println(*runway)
	}

	fmt.Println(runwayBoundingBox(runways))
}
