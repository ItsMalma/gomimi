package gomimi

import (
	"fmt"
	"strings"
)

func writeColumnPostgreSQL(column ColumnDefinition) string {
	queryBuilder := new(strings.Builder)

	queryBuilder.WriteString(fmt.Sprintf(`"%v" %v`, column.Name, column.Type))
	if column.Default != "" {
		queryBuilder.WriteString(fmt.Sprintf(` DEFAULT %v`, column.Default))
	}
	if column.Nullable {
		queryBuilder.WriteString(fmt.Sprintf(` NULL`))
	} else {
		queryBuilder.WriteString(fmt.Sprintf(` NOT NULL`))
	}
	if column.PrimaryKey {
		queryBuilder.WriteString(fmt.Sprintf(` PRIMARY KEY`))
	}
	if column.Unique {
		queryBuilder.WriteString(fmt.Sprintf(` UNIQUE`))
	}
	if column.Reference {
		queryBuilder.WriteString(fmt.Sprintf(` REFERENCES "%v" (`, column.ReferenceTableName))
		referenceColumnNamesLength := len(column.ReferenceColumnNames)
		for index2, referenceColumnName := range column.ReferenceColumnNames {
			queryBuilder.WriteString(fmt.Sprintf(`"%v"`, referenceColumnName))
			if index2+1 < referenceColumnNamesLength {
				queryBuilder.WriteString(`,`)
			}
			queryBuilder.WriteString(`)`)
		}
	}
	if column.CheckExpression != "" {
		queryBuilder.WriteString(fmt.Sprintf(` CHECK (%v)`, column.CheckExpression))
	}
	if column.AutoIncrement {
		queryBuilder.WriteString(fmt.Sprintf(` GENERATED ALWAYS AS IDENTITY`))
	}

	return queryBuilder.String()
}

func writeConstraintPostgreSQL(constraint ConstraintDefinition) string {
	queryBuilder := new(strings.Builder)

	if constraint.Name != "" {
		queryBuilder.WriteString(fmt.Sprintf(`CONSTRAINT "%v"`, constraint.Name))
	}
	switch constraint.Type {
	case ConstraintPrimaryKey:
		queryBuilder.WriteString(` PRIMARY KEY (`)
		columnNamesLength := len(constraint.ColumnNames)
		for index2, columnName := range constraint.ColumnNames {
			queryBuilder.WriteString(fmt.Sprintf(`"%v"`, columnName))
			if index2+1 < columnNamesLength {
				queryBuilder.WriteString(`,`)
			}
		}
		queryBuilder.WriteString(`)`)
	case ConstraintUnique:
		queryBuilder.WriteString(` UNIQUE (`)
		columnNamesLength := len(constraint.ColumnNames)
		for index2, columnName := range constraint.ColumnNames {
			queryBuilder.WriteString(fmt.Sprintf(`"%v"`, columnName))
			if index2+1 < columnNamesLength {
				queryBuilder.WriteString(`,`)
			}
		}
		queryBuilder.WriteString(`)`)
	case ConstraintForeignKey:
		queryBuilder.WriteString(` FOREIGN KEY (`)
		columnNamesLength := len(constraint.ColumnNames)
		for index2, columnName := range constraint.ColumnNames {
			queryBuilder.WriteString(fmt.Sprintf(`"%v"`, columnName))
			if index2+1 < columnNamesLength {
				queryBuilder.WriteString(`,`)
			}
		}
		queryBuilder.WriteString(`)`)
		queryBuilder.WriteString(fmt.Sprintf(` REFERENCES "%v" (`, constraint.ReferenceTableName))
		referenceColumnNamesLength := len(constraint.ReferenceColumnNames)
		for index2, referenceColumnName := range constraint.ReferenceColumnNames {
			queryBuilder.WriteString(fmt.Sprintf(`"%v"`, referenceColumnName))
			if index2+1 < referenceColumnNamesLength {
				queryBuilder.WriteString(`,`)
			}
			queryBuilder.WriteString(`)`)
		}
	case ConstraintCheck:
		queryBuilder.WriteString(fmt.Sprintf(` CHECK (%v)`, constraint.CheckExpression))
	}

	return queryBuilder.String()
}

type builderPostgreSQL struct {
	queryBuilder *strings.Builder
}

func NewBuilderPostgreSQL() Builder {
	return &builderPostgreSQL{queryBuilder: new(strings.Builder)}
}

func (builder *builderPostgreSQL) Begin() {
	builder.queryBuilder.WriteString("BEGIN;\n\n")
}

func (builder *builderPostgreSQL) Rollback() {
	builder.queryBuilder.WriteString("ROLLBACK;")
	builder.queryBuilder.Reset()
}

func (builder *builderPostgreSQL) Commit() string {
	builder.queryBuilder.WriteString("COMMIT;")
	result := builder.queryBuilder.String()
	builder.queryBuilder.Reset()
	return result
}

func (builder *builderPostgreSQL) CreateTable(name string, columns []ColumnDefinition, constraints []ConstraintDefinition) TableBuilder {
	builder.queryBuilder.WriteString(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%v" (`, name))

	columnsLength := len(columns)
	for index, column := range columns {
		builder.queryBuilder.WriteString(fmt.Sprintf(`%v`, writeColumnPostgreSQL(column)))
		if index+1 < columnsLength {
			builder.queryBuilder.WriteString(`,`)
		}
	}

	constraintsLength := len(constraints)
	for index, constraint := range constraints {
		builder.queryBuilder.WriteString(fmt.Sprintf(`%v`, writeConstraintPostgreSQL(constraint)))
		if index+1 < constraintsLength {
			builder.queryBuilder.WriteString(`,`)
		}
	}

	builder.queryBuilder.WriteString(`);\n\n`)

	return &tableBuilderPostgreSQL{tableName: name, queryBuilder: builder.queryBuilder}
}

func (builder *builderPostgreSQL) AlterTable(name string) TableBuilder {
	return &tableBuilderPostgreSQL{tableName: name, queryBuilder: builder.queryBuilder}
}

func (builder *builderPostgreSQL) DropTable(name string) Builder {
	builder.queryBuilder.WriteString(fmt.Sprintf(`DROP TABLE IF EXISTS "%v";\n\n`, name))
	return builder
}

func (builder *builderPostgreSQL) TruncateTable(name string) Builder {
	builder.queryBuilder.WriteString(fmt.Sprintf(`TRUNCATE TABLE "%v";\n\n`, name))
	return builder
}

type tableBuilderPostgreSQL struct {
	tableName    string
	queryBuilder *strings.Builder
}

func (builder *tableBuilderPostgreSQL) Rename(newTableName string) TableBuilder {
	builder.queryBuilder.WriteString(fmt.Sprintf(`ALTER TABLE IF EXISTS "%v" RENAME TO "%v";\n\n`, builder.tableName, newTableName))
	return builder
}

func (builder *tableBuilderPostgreSQL) AddColumn(column ColumnDefinition) TableBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ADD COLUMN IF NOT EXISTS %v;\n\n`,
			builder.tableName,
			writeColumnPostgreSQL(column),
		),
	)

	return builder
}

func (builder *tableBuilderPostgreSQL) AddConstraint(constraint ConstraintDefinition) TableBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ADD %v;\n\n`,
			builder.tableName,
			writeConstraintPostgreSQL(constraint),
		),
	)

	return builder
}

func (builder *tableBuilderPostgreSQL) AddIndex(index IndexDefinition) TableBuilder {
	builder.queryBuilder.WriteString(`CREATE `)
	if index.Unique {
		builder.queryBuilder.WriteString(`UNIQUE `)
	}
	builder.queryBuilder.WriteString(`INDEX `)
	if index.Name != "" {
		builder.queryBuilder.WriteString(fmt.Sprintf(`IF NOT EXISTS "%v" `, index.Name))
	}

	builder.queryBuilder.WriteString(fmt.Sprintf(`ON "%v" (`, builder.tableName))
	columnNamesLength := len(index.ColumnNames)
	for i, columnName := range index.ColumnNames {
		builder.queryBuilder.WriteString(fmt.Sprintf(`"%v"`, columnName))
		if i+1 < columnNamesLength {
			builder.queryBuilder.WriteString(`,`)
		}
	}
	builder.queryBuilder.WriteString(`)`)

	if index.OnExpression != "" {
		builder.queryBuilder.WriteString(fmt.Sprintf(` WHERE %v`, index.OnExpression))
	}

	builder.queryBuilder.WriteString(`\n\n`)

	return builder
}

func (builder *tableBuilderPostgreSQL) AlterColumn(columnName string, callback func(alterColumnBuilder AlterColumnBuilder)) TableBuilder {
	callback(&alterColumnBuilderPostgreSQL{tableName: builder.tableName, columnName: columnName, queryBuilder: builder.queryBuilder})
	return builder
}

func (builder *tableBuilderPostgreSQL) DropColumn(columnName string) TableBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" DROP COLUMN IF EXISTS "%v";\n\n`,
			builder.tableName,
			columnName,
		),
	)

	return builder
}

func (builder *tableBuilderPostgreSQL) DropConstraint(constraintName string) TableBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" DROP CONSTRAINT IF EXISTS "%v";\n\n`,
			builder.tableName,
			constraintName,
		),
	)

	return builder
}

func (builder *tableBuilderPostgreSQL) DropIndex(indexName string) TableBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`DROP INDEX IF EXISTS "%v";\n\n`,
			indexName,
		),
	)

	return builder
}

func (builder *tableBuilderPostgreSQL) RenameColumn(oldColumnName string, newColumnName string) TableBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" RENAME COLUMN "%v" TO "%v";\n\n`,
			builder.tableName,
			oldColumnName,
			newColumnName,
		),
	)

	return builder
}

func (builder *tableBuilderPostgreSQL) RenameConstraint(oldConstraintName string, newConstraintName string) TableBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" RENAME CONSTRAINT "%v" TO "%v";\n\n`,
			builder.tableName,
			oldConstraintName,
			newConstraintName,
		),
	)

	return builder
}

func (builder *tableBuilderPostgreSQL) RenameIndex(oldIndexName string, newIndexName string) TableBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER INDEX IF EXISTS "%v" RENAME TO "%v";\n\n`,
			oldIndexName,
			newIndexName,
		),
	)

	return builder
}

type columnBuilderPostgreSQL struct {
	queryBuilder *strings.Builder

	definition ColumnDefinition
}

func (builder *columnBuilderPostgreSQL) WithName(name string) ColumnBuilder {
	builder.definition.Name = name
	return builder
}

func (builder *columnBuilderPostgreSQL) WithType(typeName string) ColumnBuilder {
	builder.definition.Type = typeName
	return builder
}

func (builder *columnBuilderPostgreSQL) WithDefault(expression string) ColumnBuilder {
	builder.definition.Default = expression
	return builder
}

func (builder *columnBuilderPostgreSQL) IsNullable(enableNullable bool) ColumnBuilder {
	builder.definition.Nullable = enableNullable
	return builder
}

func (builder *columnBuilderPostgreSQL) IsPrimaryKey(enablePrimaryKey bool) ColumnBuilder {
	builder.definition.PrimaryKey = enablePrimaryKey
	return builder
}

func (builder *columnBuilderPostgreSQL) IsUnique(enableUnique bool) ColumnBuilder {
	builder.definition.Unique = enableUnique
	return builder
}

func (builder *columnBuilderPostgreSQL) IsForeignKey(enableForeign bool, referenceTableName string, referenceColumnNames ...string) ColumnBuilder {
	builder.definition.Reference = enableForeign
	builder.definition.ReferenceTableName = referenceTableName
	builder.definition.ReferenceColumnNames = referenceColumnNames
	return builder
}

func (builder *columnBuilderPostgreSQL) IsCheck(expression string) ColumnBuilder {
	builder.definition.CheckExpression = expression
	return builder
}

func (builder *columnBuilderPostgreSQL) IsAutoIncrement(enableAutoIncrement bool) ColumnBuilder {
	builder.definition.AutoIncrement = enableAutoIncrement
	return builder
}

func (builder *columnBuilderPostgreSQL) Build() ColumnDefinition {
	return builder.definition
}

type alterColumnBuilderPostgreSQL struct {
	tableName    string
	columnName   string
	queryBuilder *strings.Builder
}

func (builder *alterColumnBuilderPostgreSQL) AlterType(typeName string) AlterColumnBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ALTER COLUMN "%v" TYPE %v;\n\n`,
			builder.tableName,
			builder.columnName,
			typeName,
		),
	)
	return builder
}

func (builder *alterColumnBuilderPostgreSQL) AlterDefault(expression string) AlterColumnBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ALTER COLUMN "%v" SET DEFAULT %v;\n\n`,
			builder.tableName,
			builder.columnName,
			expression,
		),
	)
	return builder
}

func (builder *alterColumnBuilderPostgreSQL) DropDefault() AlterColumnBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ALTER COLUMN "%v" DROP DEFAULT;\n\n`,
			builder.tableName,
			builder.columnName,
		),
	)
	return builder
}

func (builder *alterColumnBuilderPostgreSQL) SetNullable() AlterColumnBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ALTER COLUMN "%v" DROP NOT NULL;\n\n`,
			builder.tableName,
			builder.columnName,
		),
	)
	return builder
}

func (builder *alterColumnBuilderPostgreSQL) DropNullable() AlterColumnBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ALTER COLUMN "%v" SET NOT NULL;\n\n`,
			builder.tableName,
			builder.columnName,
		),
	)
	return builder
}

func (builder *alterColumnBuilderPostgreSQL) SetAutoIncrement() AlterColumnBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ALTER COLUMN "%v" ADD GENERATED ALWAYS AS IDENTITY;\n\n`,
			builder.tableName,
			builder.columnName,
		),
	)
	return builder
}

func (builder *alterColumnBuilderPostgreSQL) DropAutoIncrement() AlterColumnBuilder {
	builder.queryBuilder.WriteString(
		fmt.Sprintf(
			`ALTER TABLE IF EXISTS "%v" ALTER COLUMN "%v" DROP IDENTITY IF EXISTS;\n\n`,
			builder.tableName,
			builder.columnName,
		),
	)
	return builder
}
