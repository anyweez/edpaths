package structs

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"

	"github.com/anyweez/edpaths/structs/gen"
	// "github.com/boltdb/bolt"
	// "github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
)

type SystemID int

/**
 * SpaceStation is a full representation of an individual station. Currently this only contains
 * navigational data (nothing commercial). It also contains the ID (key) of the system its in.
 */
type SpaceStation struct {
	ID             SystemID `json:"id"`
	Name           string
	DistanceToStar int
	SystemID       int
}

/**
 * SpaceStation is a full representation of an individual system. Currently this only contains
 * navigational data (nothing commercial).
 */
type SpaceSystem struct {
	ID   SystemID `json:"id"`
	Name string
	X    float64
	Y    float64
	Z    float64

	Bucket                *SpaceBucket `json:"-"`
	ContainsScoopableStar bool
	ContainsRefuelStation bool
}

type SpaceBody struct {
	ID       int
	GroupID  int      `json:"group_id"` // 6 = star
	SystemID SystemID `json:"system_id"`

	SpectralClass string `json:"spectral_class"`
}

type SpaceStop struct {
	System           *SpaceSystem
	DistanceFromPrev float64
	RequestedStop    bool
	// ID                    SystemID
	// Name                  string
	// ContainsScoopableStar bool
	// ContainsRefuelStation bool
}

type SpaceDB struct {
	Stations []*SpaceSystem
	Systems  []*SpaceSystem
}

func Connect(dbPath string) *SpaceDB {
	fmt.Println("Connecting to SpaceDB")
	var err error

	db := new(SpaceDB)
	systems, _ := ioutil.ReadFile("data/" + dbPath + ".db")

	universe := space.Universe{}
	proto.Unmarshal(systems, &universe)

	db.Systems = make([]*SpaceSystem, len(universe.GetSystems()))

	for i, sys := range universe.GetSystems() {
		// Space is actually allocated for all systems here, and only here. Any other
		// data structure should maintain a reference to this object.
		db.Systems[i] = &SpaceSystem{
			ID:   SystemID(sys.GetSystemID()),
			Name: sys.GetName(),
			X:    sys.GetX(),
			Y:    sys.GetY(),
			Z:    sys.GetZ(),
			ContainsRefuelStation: sys.GetContainsRefuelStation(),
			ContainsScoopableStar: sys.GetContainsScoopableStar(),
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	return db
}

func (db *SpaceDB) ForEachSystem(each func(*SpaceSystem)) {
	for _, system := range db.Systems {
		each(system)
	}
}

func (src *SpaceSystem) DistanceTo(dest *SpaceSystem) float64 {
	return math.Sqrt((dest.X-src.X)*(dest.X-src.X) + (dest.Y-src.Y)*(dest.Y-src.Y) + (dest.Z-src.Z)*(dest.Z-src.Z))
}

func (src *SpaceSystem) AsStop() *SpaceStop {
	return &SpaceStop{
		System: src,
	}
}
