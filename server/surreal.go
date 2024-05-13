package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/conn/gorilla"
	"github.com/surrealdb/surrealdb.go/pkg/marshal"
)

var surrealLock = &sync.Mutex{}

var surreal *surrealdb.DB

func Surreal() *surrealdb.DB {
	return surreal
}

func SurrealInit() error {
	surrealLock.Lock()
	defer surrealLock.Unlock()

	if surreal == nil {

		url := os.Getenv("SURREAL_URL")
		username := os.Getenv("SURREAL_USERNAME")
		password := os.Getenv("SURREAL_PASSWORD")
		database := os.Getenv("SURREAL_DATABASE")
		namespace := os.Getenv("SURREAL_NAMESPACE")

		db, err := surrealdb.New(url, gorilla.Create())
		if err != nil {
			return err
		}

		if _, err = db.Use(namespace, database); err != nil {
			return err
		}

		authData := &surrealdb.Auth{
			Username:  username,
			Password:  password,
			Namespace: namespace,
		}
		if _, err = db.Signin(authData); err != nil {
			return err
		}

		surreal = db
	}

	return nil
}

type Point struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Site struct {
	ID        string    `json:"id"`
	ICAO      string    `json:"icao"`
	Name      string    `json:"name"`
	LastScan  time.Time `json:"last_scan"`
	Type      string    `json:"type"`
	VCP       int       `json:"vcp"`
	Elevation int       `json:"elevation"`
	Location  Point     `json:"location"`
}

func GetSite(icao string) (*Site, error) {
	result, err := Surreal().Query(fmt.Sprintf("SELECT * FROM radar_site:%s", icao), map[string]string{})
	if err != nil {
		return nil, err
	}

	// NOTE: Surreal returns an array of the result which requires an array to be Unmarshalled. This is referenced later
	sites := make([]Site, 1)
	err = marshal.Unmarshal(result, &sites)
	if err != nil {
		return nil, err
	}

	if len(sites) != 0 {
		return &sites[0], nil
	} else {
		return nil, nil
	}
}

func AddSite(icao string) (*Site, error) {

	site := Site{
		ID:        icao,
		ICAO:      icao,
		Name:      "",
		LastScan:  time.Now().UTC(),
		Type:      "",
		VCP:       0,
		Elevation: 0,
	}

	data, err := surreal.Create("radar_site", site)
	if err != nil {
		return nil, err
	}

	newSite := make([]Site, 1)
	err = marshal.Unmarshal(data, &newSite)
	if err != nil {
		return nil, err
	}

	return &newSite[0], nil

}

func GetVolume(id string) (*Volume, error) {
	result, err := Surreal().Query(fmt.Sprintf("SELECT * FROM radar_volume:%s", id), map[string]string{})
	if err != nil {
		return nil, err
	}

	// NOTE: Surreal returns an array of the result which requires an array to be Unmarshalled. This is referenced later
	volumes := make([]Volume, 1)
	err = marshal.Unmarshal(result, &volumes)
	if err != nil {
		return nil, err
	}

	if len(volumes) != 0 {
		return &volumes[0], nil
	} else {
		return nil, nil
	}
}
