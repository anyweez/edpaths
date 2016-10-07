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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/anyweez/edpaths/structs"
	"github.com/anyweez/edpaths/structs/gen"
	"github.com/golang/protobuf/proto"
)

/**
 * Reads in all station data and adds each record to a Bolt database keyed by station ID.
 * Also outputs a CSV containing all station names and their ID.
 */
// func stations() {
// 	fp, _ := os.Open("data/stations.json")
// 	decoder := json.NewDecoder(fp)

// 	db, _ := bolt.Open("data/stations.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
// 	db.Update(func(tx *bolt.Tx) error {
// 		tx.CreateBucketIfNotExists([]byte("stations"))
// 		return nil
// 	})
// 	defer db.Close()

// 	out, _ := os.Create("data/stations.csv")
// 	csv := csv.NewWriter(out)
// 	csv.Write([]string{"name", "id"})

// 	decoder.Token()
// 	for decoder.More() {
// 		var station structs.SpaceStation

// 		if err := decoder.Decode(&station); err != nil {
// 			log.Fatal(err)
// 		}

// 		// Update the database
// 		db.Update(func(tx *bolt.Tx) error {
// 			b := tx.Bucket([]byte("stations"))

// 			raw := new(bytes.Buffer)
// 			gob.NewEncoder(raw).Encode(station)
// 			b.Put([]byte(strconv.Itoa(int(station.ID))), raw.Bytes())

// 			return nil
// 		})

// 		// Update the CSV
// 		csv.Write([]string{station.Name, strconv.Itoa(int(station.ID))})
// 	}
// 	decoder.Token()
// 	csv.Flush()

// 	if err := csv.Error(); err != nil {
// 		log.Fatal(err)
// 	}
// }

func systems(in chan structs.SpaceSystem, status *sync.WaitGroup) {
	// Set up (new) or load (existing) database and prepare to make some changes.
	full := space.Universe{}
	sample := space.Universe{}

	centroid := structs.SpaceSystem{X: 100, Y: 100, Z: 100}

	for system := range in {
		id := int32(system.ID)
		nextSystem := space.SpaceSystem{
			SystemID: &id,
			Name:     &system.Name,
			X:        &system.X,
			Y:        &system.Y,
			Z:        &system.Z,
			ContainsScoopableStar: &system.ContainsScoopableStar,
			ContainsRefuelStation: &system.ContainsRefuelStation,
		}

		full.Systems = append(full.Systems, &nextSystem)

		// Update the sample DB iff its within 100 LY of the definied centroid
		if centroid.DistanceTo(&system) < 100 {
			sample.Systems = append(sample.Systems, &nextSystem)
		}
	}

	fullOut, _ := proto.Marshal(&full)
	sampleOut, _ := proto.Marshal(&sample)

	ioutil.WriteFile("data/systems.db", fullOut, 0644)
	ioutil.WriteFile("data/sample.db", sampleOut, 0644)

	status.Done()
}

func main() {
	var status sync.WaitGroup
	status.Add(1)

	sys := make(chan structs.SpaceSystem, 100)

	fmt.Println("Reading system data...")
	go LoadSystems(sys)
	go systems(sys, &status)

	status.Wait()
}

/**
 * Read all systems and bodies and push them out into the provided channel once they're
 * available.
 */
func LoadSystems(out chan structs.SpaceSystem) {
	// Load bodies first and generate the `powered` map
	powered := make(map[structs.SystemID]bool)

	scoopable := []string{"O", "B", "A", "F", "G", "K", "M"}
	bfp, _ := os.Open("data/bodies.json")
	bDecoder := json.NewDecoder(bfp)

	bDecoder.Token()
	for bDecoder.More() {
		var body structs.SpaceBody

		if err := bDecoder.Decode(&body); err != nil {
			log.Fatal(err)
		}

		// If we're dealing with a star, check to see whether its scoopable
		powered[body.SystemID] = false

		if body.GroupID == 6 {
			for _, class := range scoopable {
				if body.SpectralClass == class {
					powered[body.SystemID] = true
				}
			}
		}
	}
	bDecoder.Token()

	// Load systems
	fp, _ := os.Open("data/systems.json")
	decoder := json.NewDecoder(fp)

	decoder.Token()
	for decoder.More() {
		var system structs.SpaceSystem

		if err := decoder.Decode(&system); err != nil {
			log.Fatal(err)
		}

		if status, exists := powered[system.ID]; exists {
			system.ContainsScoopableStar = status
		} else {
			system.ContainsScoopableStar = false
		}

		out <- system
	}

	decoder.Token()
	close(out)
}
