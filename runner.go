package gomimi

import "fmt"

type Runner struct {
	indicator Indicator
	builder   Builder
}

func NewRunner(indicator Indicator, builder Builder) Runner {
	return Runner{indicator, builder}
}

func (runner Runner) RunMigration(migrations ...Migration) {
	currentMigrationName := runner.indicator.Current()
	notFound := true

	for _, migration := range migrations {
		if !notFound {
			runner.builder.Begin()
			if err := migration.Up(runner.builder); err != nil {
				runner.builder.Rollback()
				runner.builder.Begin()
				if err := migration.Down(runner.builder); err != nil {
					runner.builder.Rollback()
					panic(err)
				}
			}
			runner.builder.Commit()
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
