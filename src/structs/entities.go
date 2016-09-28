package structs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"strconv"

	"github.com/boltdb/bolt"
)

type SystemID int

/**
 * SpaceStation is a full representation of an individual station. Currently this only contains
 * navigational data (nothing commercial). It also contains the ID (key) of the system its in.
 */
type SpaceStation struct {
	ID             SystemID
	Name           string
	DistanceToStar int
	SystemID       int
}

/**
 * SpaceStation is a full representation of an individual system. Currently this only contains
 * navigational data (nothing commercial).
 */
type SpaceSystem struct {
	ID   SystemID
	Name string
	X    float64
	Y    float64
	Z    float64

	Bucket *SpaceBucket
}

type SpaceDB struct {
	Stations *bolt.DB
	Systems  *bolt.DB
}

func Connect() *SpaceDB {
	fmt.Println("Connecting to SpaceDB")
	db := new(SpaceDB)
	db.Stations, _ = bolt.Open("data/stations.db", 0400, nil)
	// TODO: revert to systems.db when the time is right
	db.Systems, _ = bolt.Open("data/sample.db", 0400, nil)

	return db
}

// systemDb := bolt.Open("data/systems.db", 0400, nil)

func (db *SpaceDB) GetSystem(id int) SpaceSystem {
	var system SpaceSystem
	var raw []byte

	db.Systems.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("systems"))
		v := bucket.Get([]byte(strconv.Itoa(id)))

		// Resize the raw array and copy data in
		raw = make([]byte, len(v))
		copy(raw, v)

		return nil
	})

	gob.NewDecoder(bytes.NewReader(raw)).Decode(&system)
	return system
}

func (db *SpaceDB) ForEachSystem(each func(SpaceSystem)) {
	db.Systems.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("systems"))
		all := bucket.Cursor()

		for k, v := all.First(); k != nil; k, v = all.Next() {
			var system SpaceSystem
			gob.NewDecoder(bytes.NewReader(v)).Decode(&system)

			each(system)
		}

		return nil
	})
}

func (src *SpaceSystem) DistanceTo(dest *SpaceSystem) float64 {
	return math.Sqrt((dest.X-src.X)*(dest.X-src.X) + (dest.Y-src.Y)*(dest.Y-src.Y) + (dest.Z-src.Z)*(dest.Z-src.Z))
}
