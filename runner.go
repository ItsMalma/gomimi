package gomimi

type Runner struct {
	builder Builder
}

func NewRunner(builder Builder) Runner {
	return Runner{builder}
}

func (runner Runner) Run(migrations ...Migration) {
	for _, migration := range migrations {
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
	}
}
