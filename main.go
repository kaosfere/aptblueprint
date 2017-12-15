package main

import "fmt"
import _ "math"
import "encoding/csv"
import bolt "github.com/coreos/bbolt"
import "os"
import "io"
import "strconv"
import _ "bytes"
import _ "encoding/gob"
import "time"
import "github.com/vmihailenco/msgpack"


type airport struct {
	Code      string
	Name      string
	Latitude  float64
	Longitude float64
	Elevation int64
	Country    string
	Region    string
	City      string
	Iata      string
}

func initDb() error {
	db, err := bolt.Open("aptdata.db", 0644, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	airportFile, err := os.Open("airports.csv")
	if err != nil {
		return err
	}
	defer airportFile.Close()

	airportReader := csv.NewReader(airportFile)
	_, err = airportReader.Read()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Airports"))
		if err != nil {
			return err
		}
		return nil})



	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Airports"))

		for {
			record, err := airportReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
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
//			buf := bytes.Buffer{}
//			e := gob.NewEncoder(&buf)
//			err = e.Encode(apt)
//			if err != nil {
//				panic(err)
//			}
//			if record[1] == "KORD" {
//				fmt.Println(apt)
//				fmt.Println(buf.Bytes())
//			}
//			err = b.Put([]byte(record[1]), buf.Bytes())
			m, err := msgpack.Marshal(&apt)
			if err != nil {
				panic(err)
			}
			err = b.Put([]byte(record[1]), m)
			if err != nil {
				panic(err)
			}
		}
		return nil})
	return nil
	}

func main() {
	start := time.Now()
	err := initDb()
	elapsed := time.Since(start)
	fmt.Println(elapsed)
	if err != nil {
		fmt.Println("ERROR", err)
	}
//	fmt.Println(err)
//	fmt.Println("Done")
	var apt airport
	db, err := bolt.Open("aptdata.db", 0644, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Airports"))
		aptg := b.Get([]byte("KORD"))
		//buf := bytes.NewBuffer(aptg)
//		buf := bytes.Buffer{}
//		buf.Write(aptg)
		//fmt.Println(buf)
//		d := gob.NewDecoder(buf)
//		err = d.Decode(&apt)
		msgpack.Unmarshal(aptg, &apt)
		fmt.Println(err)
		fmt.Println(apt)
		return nil
		})

}
