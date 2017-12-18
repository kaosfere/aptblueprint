package main

import "github.com/pkg/errors"
import "fmt"
import "github.com/coreos/bbolt"
import "github.com/vmihailenco/msgpack"
import "os"

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

func filterForCoords(raw []*runway) []*runway {
	filtered := []*runway{}
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
	db, err := buildDB(false)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	runways, err := getRunways(db, "KDEN")
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
