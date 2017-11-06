package database

import (
	"log"
)

const (
	minimalVersion = "1.0"
	latestVersion  = "1.0"
)

type dbUpdater struct {
	version  string
	updateDb func(db *Database)
}

func UpdateVersion(db *Database) {
	currentVersion := db.GetDatabaseVersion()

	if currentVersion != latestVersion {
		updaters := makeUpdaters(currentVersion, latestVersion)

		for _, updater := range updaters {
			updater.updateDb(db)
		}
		log.Printf("Update DB version from %s to %s", currentVersion, latestVersion)
	}

	db.SetDatabaseVersion(latestVersion)
}

func makeUpdaters(versionFrom string, versionTo string) (updaters []dbUpdater) {
	allUpdaters := makeAllUpdaters()

	isFirstFound := (versionFrom == minimalVersion)
	for _, updater := range allUpdaters {
		if isFirstFound {
			updaters = append(updaters, updater)
			if updater.version == versionTo {
				break
			}
		} else {
			if updater.version == versionFrom {
				isFirstFound = true
				updaters = append(updaters, updater)
			}
		}
	}
	return
}

func makeAllUpdaters() (updaters []dbUpdater) {
	updaters = []dbUpdater{
		dbUpdater{
			// 1.0 doesn't have version field, so you should add it manually
			version: "1.2",
			updateDb: func(db *Database) {
				db.execQuery("ALTER TABLE users ADD COLUMN banned")
			},
		},
	}
	return
}
