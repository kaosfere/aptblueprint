package main

import "github.com/pkg/errors"
import "fmt"
import "github.com/coreos/bbolt"
import "github.com/vmihailenco/msgpack"
import "os"

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
	}

	err = loadAirports(db)
	if err != nil {
		return nil, errors.Wrap(err, "loading airports")
	}
	err = loadRunways(db)
	if err != nil {
		return nil, errors.Wrap(err, "loading runways")
	}
	return db, err
}

func main() {
	var apt airport
	db, err := buildDB(false)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	err = db.View(func(tx *bolt.Tx) error {
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
		/*
			fmt.Println("sub bucket", wayg)
			msgpack.Unmarshal(wayg, &rwy)
			fmt.Println(rwy)*/
		return nil
	})
	if err != nil {
		fmt.Println("POOP", err)
	}

}
