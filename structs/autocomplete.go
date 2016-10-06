package structs

import "strings"

type SystemRecord struct {
	Name string
	ID   SystemID
}

type Autocomplete struct {
	records []*SystemRecord
}

func NewAutocomplete(db *SpaceDB) Autocomplete {
	ac := Autocomplete{
		records: make([]*SystemRecord, 0),
	}

	// Populate the autocorrecter
	db.ForEachSystem(func(system SpaceSystem) {
		ac.Add(&SystemRecord{
			Name: system.Name,
			ID:   system.ID,
		})
	})

	return ac
}

/**
 * Add a new system record. Will be returned immediately
 */
func (ac *Autocomplete) Add(record *SystemRecord) {
	ac.records = append(ac.records, record)
}

/**
 * Returns up to LIMIT records that match the provided string.
 */
func (ac *Autocomplete) GetAll(fragment string, limit int) []*SystemRecord {
	results := make([]*SystemRecord, 0, limit)

	// Very slow implementation currently
	for _, record := range ac.records {
		if strings.Contains(strings.ToLower(record.Name), strings.ToLower(fragment)) {
			results = append(results, record)
		}

		// Call it quits early if we reach the cap
		if len(results) == limit {
			return results
		}
	}

	return results
}
