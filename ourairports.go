package main

import "github.com/pkg/errors"
import "fmt"
import "encoding/csv"
import "github.com/coreos/bbolt"
import "os"
import "io"
import "strconv"
import "github.com/vmihailenco/msgpack"

type airport struct {
	Code      string
	Name      string
	Latitude  float64
	Longitude float64
	Elevation int64
	Country   string
	Region    string
	City      string
	Iata      string
}

type runway struct {
	Airport       string
	Length        int64
	Width         int64
	Surface       string
	Lighted       bool
	Closed        bool
	End1Name      string
	End1Latitude  float64
	End1Longitude float64
	End1Elevation int64
	End1Heading   int64
	End1Displaced int64
	End2Name      string
	End2Latitude  float64
	End2Longitude float64
	End2Elevation int64
	End2Heading   int64
	End2Displaced int64
}

func loadAirports(db *bolt.DB) error {
	apts, err := os.Open("airports.csv")
	if err != nil {
		return err
	}
	defer apts.Close()

	r := csv.NewReader(apts)
	_, err = r.Read() // skip header

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("Airports"))
		if err != nil {
			return err
		}
		b := tx.Bucket([]byte("Airports"))

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return errors.Wrap(err, "airport read")
			}

			latitude, _ := strconv.ParseFloat(record[4], 64)
			longitude, _ := strconv.ParseFloat(record[5], 64)
			elevation, _ := strconv.ParseInt(record[6], 10, 64)
			apt := airport{record[1],
				record[3],
				latitude,
				longitude,
				elevation,
				record[8],
				record[9],
				record[10],
				record[13]}

			m, err := msgpack.Marshal(&apt)
			if err != nil {
				return errors.Wrap(err, "airport marshal")
			}

			err = b.Put([]byte(record[1]), m)
			if err != nil {
				return errors.Wrap(err, "database put")
			}

		}

		return nil
	})

	return err
}

func loadRunways(db *bolt.DB) error {
	rwys, err := os.Open("runways.csv")
	if err != nil {
		return err
	}
	defer rwys.Close()

	r := csv.NewReader(rwys)
	r.FieldsPerRecord = -1 // extra comma on first line
	_, err = r.Read()      // skip header

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Runways"))
		if err != nil {
			fmt.Println(err)
			return err
		}
		//b := tx.Bucket([]byte("Runways"))

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				return errors.Wrap(err, "runway read")
			}

			length, _ := strconv.ParseInt(record[3], 10, 64)
			width, _ := strconv.ParseInt(record[4], 10, 64)
			lighted := record[6] == "1"
			closed := record[7] == "1"
			end1Latitude, _ := strconv.ParseFloat(record[9], 64)
			end1Longitude, _ := strconv.ParseFloat(record[10], 64)
			end1Elevation, _ := strconv.ParseInt(record[11], 10, 64)
			end1Heading, _ := strconv.ParseInt(record[12], 10, 64)
			end1Displaced, _ := strconv.ParseInt(record[13], 10, 64)
			end2Latitude, _ := strconv.ParseFloat(record[15], 64)
			end2Longitude, _ := strconv.ParseFloat(record[16], 64)
			end2Elevation, _ := strconv.ParseInt(record[17], 10, 64)
			end2Heading, _ := strconv.ParseInt(record[18], 10, 64)
			end2Displaced, _ := strconv.ParseInt(record[19], 10, 64)

			rwy := runway{record[2],
				length,
				width,
				record[5],
				lighted,
				closed,
				record[8],
				end1Latitude,
				end1Longitude,
				end1Heading,
				end1Elevation,
				end1Displaced,
				record[14],
				end2Latitude,
				end2Longitude,
				end2Elevation,
				end2Heading,
				end2Displaced}

			m, err := msgpack.Marshal(&rwy)
			if err != nil {
				return errors.Wrap(err, "runway marshal")
			}

			b2, err := b.CreateBucketIfNotExists([]byte(record[2]))
			if err != nil {
				return errors.Wrap(err, "bucket creation")
			}
			err = b2.Put([]byte(record[8]+"/"+record[14]), m)
			if err != nil {
				return errors.Wrap(err, "database put")
			}

		}

		return nil
	})

	return err
}
