package gomimi

import (
	"database/sql"
	"errors"
)

type indicatorPostgreSQL struct {
	db *sql.DB
}

func NewIndicatorPostgreSQL(db *sql.DB) Indicator {
	return &indicatorPostgreSQL{db}
}

func (indicator *indicatorPostgreSQL) IfTableExists() bool {
	row := indicator.db.QueryRow(`SELECT EXISTS (SELECT FROM "pg_tables" WHERE tablename = 'gomimi');`)
	if err := row.Err(); err != nil {
		panic(err)
	}
	var exists bool
	if err := row.Scan(&exists); err != nil {
		panic(err)
	}
	return exists
}

func (indicator *indicatorPostgreSQL) CreateMigrationTable() {
	_, err := indicator.db.Exec(`CREATE TABLE "gomimi" (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
		name TEXT NOT NULL
	);`)
	if err != nil {
		panic(err)
	}
}

func (indicator *indicatorPostgreSQL) Current() string {
	if !indicator.IfTableExists() {
		return ""
	}

	row := indicator.db.QueryRow(`SELECT "name" FROM "gomimi" LIMIT 1;`)
	if err := row.Err(); err != nil {
		panic(err)
	}

	var currentName string
	if err := row.Scan(&currentName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			currentName = ""
		} else {
			panic(err)
		}
	}

	return currentName
}

func (indicator *indicatorPostgreSQL) Change(newName string) {
	if !indicator.IfTableExists() {
		indicator.CreateMigrationTable()
	}

	row := indicator.db.QueryRow(`SELECT "id" FROM "gomimi" LIMIT 1;`)
	if err := row.Err(); err != nil {
		panic(err)
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err := indicator.db.Exec(`INSERT INTO "gomimi" ("name") VALUES (?)`, newName)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		_, err := indicator.db.Exec(`UPDATE FROM "gomimi" SET "name" = ? WHERE id = ?`, newName, id)
		if err != nil {
			panic(err)
		}
	}
}
