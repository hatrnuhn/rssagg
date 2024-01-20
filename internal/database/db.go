package database

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"
)

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	err := db.ensureDB()
	if err != nil {
		return &DB{}, err
	}

	return &db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbS, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbS.Chirps) + 1

	req := ChirpReq{}
	err = json.Unmarshal([]byte(body), &req)
	if err != nil {
		return Chirp{}, errors.New("unmarshall error")
	}

	chp := Chirp{
		ID:   id,
		Body: req.Body,
	}

	dbS.Chirps[id] = chp
	err = db.writeDB(dbS)
	if err != nil {
		return Chirp{}, errors.New("writeDB error")
	}

	return chp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbS, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := []Chirp{}
	for _, chirp := range dbS.Chirps {
		chirps = append(chirps, chirp)
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})
	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)

	if os.IsNotExist(err) {
		file, err := os.Create(db.path)
		if err != nil {
			return err
		}
		defer file.Close()
	} else if err != nil {
		return err
	}

	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	body, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	dbS := DBStructure{}
	// If file isn't empty, unmarshal
	if len(body) != 0 {
		err = json.Unmarshal(body, &dbS)
		if err != nil {
			return DBStructure{}, errors.New("unmarshal loadDB error")
		}
	} else {
		dbS.Chirps = make(map[int]Chirp)
	}

	return dbS, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0644)
	if err != nil {
		return err
	}

	return nil
}
