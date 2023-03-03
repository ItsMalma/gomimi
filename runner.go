package gomimi

import (
	"database/sql"
	"fmt"
)

type Runner struct {
	indicator Indicator
	builder   Builder
}

func NewRunner(indicator Indicator, builder Builder) Runner {
	return Runner{indicator, builder}
}

func (runner Runner) RunMigration(db *sql.DB, migrations ...Migration) {
	currentMigrationName := runner.indicator.Current()
	notFound := true
	if currentMigrationName == "" {
		notFound = false
	}

	for _, migration := range migrations {
		if !notFound {
			// begin (for up migration)
			runner.builder.Begin()
			// run up migration
			if err := migration.Up(runner.builder); err != nil {
				// if up fail
				// do the rollback
				runner.builder.Rollback()
				// begin (for down migration)
				runner.builder.Begin()
				// run down migration
				if err := migration.Down(runner.builder); err != nil {
					// if down fail
					// do the rollback
					runner.builder.Rollback()
					// and panic the error
					panic(err)
				} else if _, err := db.Exec(runner.builder.Commit()); err != nil {
					// if down success
					// do commit and run the query
					// but if query fail
					// do rollback
					runner.builder.Rollback()
					// and also panic the error
					panic(err)
				}
			} else if _, err := db.Exec(runner.builder.Commit()); err != nil {
				// if up success
				// do commit and run the query
				// but if query fail
				// do rollback
				runner.builder.Rollback()
				// begin (for down migration)
				runner.builder.Begin()
				// run down migration
				if err := migration.Down(runner.builder); err != nil {
					// if down fail
					// do the rollback
					runner.builder.Rollback()
					// and panic the error
					panic(err)
				} else if _, err := db.Exec(runner.builder.Commit()); err != nil {
					// if down success
					// do commit and run the query
					// but if query fail
					// do rollback
					runner.builder.Rollback()
					// and also panic the error
					panic(err)
				}
			}
			currentMigrationName = migration.Name()
		} else if migration.Name() == currentMigrationName {
			notFound = false
		}
	}

	if notFound {
		panic(fmt.Errorf(`migration with name "%v" not found`, currentMigrationName))
	} else {
		runner.indicator.Change(currentMigrationName)
	}
}
