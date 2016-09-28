package main

/**
 * This application is responsible for ingesting, transforming, and outputting datasets into a form that's
 * ready for consumption by the service. It's currently reading in all systems and stations and outputting
 * them into Bolt (https://github.com/boltdb/bolt) databases as well as a simple <name>,<id> pairing in CSV's
 * for the purposes of autocomplete.
 *
 * This application should be run once whenever new data is downloaded.
 */

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"structs"
	"time"

	"github.com/boltdb/bolt"
)

/**
 * Reads in all station data and adds each record to a Bolt database keyed by station ID.
 * Also outputs a CSV containing all station names and their ID.
 */
func stations() {
	fp, _ := os.Open("data/stations.json")
	decoder := json.NewDecoder(fp)

	db, _ := bolt.Open("data/stations.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("stations"))
		return nil
	})
	defer db.Close()

	out, _ := os.Create("data/stations.csv")
	csv := csv.NewWriter(out)
	csv.Write([]string{"name", "id"})

	decoder.Token()
	for decoder.More() {
		var station structs.SpaceStation

		if err := decoder.Decode(&station); err != nil {
			log.Fatal(err)
		}

		// Update the database
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("stations"))

			raw := new(bytes.Buffer)
			gob.NewEncoder(raw).Encode(station)
			b.Put([]byte(strconv.Itoa(int(station.ID))), raw.Bytes())

			return nil
		})

		// Update the CSV
		csv.Write([]string{station.Name, strconv.Itoa(int(station.ID))})
	}
	decoder.Token()
	csv.Flush()

	if err := csv.Error(); err != nil {
		log.Fatal(err)
	}
}

func systems() {
	fp, _ := os.Open("data/systems.json")
	decoder := json.NewDecoder(fp)

	// Set up (new) or load (existing) database and prepare to make some changes.
	db, _ := bolt.Open("data/systems.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("systems"))
		return nil
	})
	defer db.Close()

	sampleDb, _ := bolt.Open("data/sample.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	sampleDb.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("systems"))
		return nil
	})
	defer sampleDb.Close()

	out, _ := os.Create("data/systems.csv")
	csv := csv.NewWriter(out)
	csv.Write([]string{"name", "id"})

	centroid := structs.SpaceSystem{X: 100, Y: 100, Z: 100}

	decoder.Token()
	for decoder.More() {
		var system structs.SpaceSystem

		if err := decoder.Decode(&system); err != nil {
			log.Fatal(err)
		}

		// Update the database
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("systems"))

			raw := new(bytes.Buffer)
			gob.NewEncoder(raw).Encode(system)

			b.Put([]byte(strconv.Itoa(int(system.ID))), raw.Bytes())

			return nil
		})

		// Update the sample DB iff its within 100 LY of the definied centroid
		if centroid.DistanceTo(&system) < 100 {
			sampleDb.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("systems"))

				raw := new(bytes.Buffer)
				gob.NewEncoder(raw).Encode(system)

				b.Put([]byte(strconv.Itoa(int(system.ID))), raw.Bytes())

				return nil
			})
		}

		// Update the CSV
		csv.Write([]string{system.Name, strconv.Itoa(int(system.ID))})
	}
	decoder.Token()
}

func main() {
	// Read JSON objects from stream.
	// As they're read, add to Bolt
	// Write <name>\t<id> to CSV
	fmt.Println("Reading station data...")
	// stations()
	fmt.Println("Reading system data...")
	systems()
}
