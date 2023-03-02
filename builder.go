package gomimi

type ColumnDefinition struct {
	Name                 string
	Type                 string
	Default              string
	Nullable             bool
	PrimaryKey           bool
	Unique               bool
	Reference            bool
	ReferenceTableName   string
	ReferenceColumnNames []string
	CheckExpression      string
	AutoIncrement        bool
}

type ConstraintDefinitionType uint8

const (
	ConstraintPrimaryKey ConstraintDefinitionType = iota
	ConstraintUnique
	ConstraintForeignKey
	ConstraintCheck
)

type ConstraintDefinition struct {
	Name                 string
	ColumnNames          []string
	Type                 ConstraintDefinitionType
	DefaultExpression    string
	ReferenceTableName   string
	ReferenceColumnNames []string
	CheckExpression      string
}

type IndexDefinition struct {
	Name         string
	ColumnNames  []string
	Unique       bool
	OnExpression string
}

type Builder interface {
	Begin()
	Rollback()
	Commit() string
	CreateTable(name string, columns []ColumnDefinition, constraints []ConstraintDefinition) TableBuilder
	AlterTable(name string) TableBuilder
	DropTable(name string) Builder
	TruncateTable(name string) Builder
}

type TableBuilder interface {
	Rename(newTableName string) TableBuilder
	AddColumn(column ColumnDefinition) TableBuilder
	AddConstraint(constraint ConstraintDefinition) TableBuilder
	AddIndex(index IndexDefinition) TableBuilder
	AlterColumn(columnName string, callback func(alterColumnBuilder AlterColumnBuilder)) TableBuilder
	DropColumn(columnName string) TableBuilder
	DropConstraint(constraintName string) TableBuilder
	DropIndex(indexName string) TableBuilder
	RenameColumn(oldColumnName string, newColumnName string) TableBuilder
	RenameConstraint(oldConstraintName string, newConstraintName string) TableBuilder
	RenameIndex(oldIndexName string, newIndexName string) TableBuilder
}

type ColumnBuilder interface {
	WithName(name string) ColumnBuilder
	WithType(typeName string) ColumnBuilder
	WithDefault(expression string) ColumnBuilder
	IsNullable(enableNullable bool) ColumnBuilder
	IsPrimaryKey(enablePrimaryKey bool) ColumnBuilder
	IsUnique(enableUnique bool) ColumnBuilder
	IsForeignKey(enableForeign bool, referenceTableName string, referenceColumnNames ...string) ColumnBuilder
	IsCheck(expression string) ColumnBuilder
	IsAutoIncrement(enableAutoIncrement bool) ColumnBuilder
	Build() ColumnDefinition
}

type AlterColumnBuilder interface {
	AlterType(typeName string) AlterColumnBuilder
	AlterDefault(expression string) AlterColumnBuilder
	DropDefault() AlterColumnBuilder
	SetNullable() AlterColumnBuilder
	DropNullable() AlterColumnBuilder
	SetAutoIncrement() AlterColumnBuilder
	DropAutoIncrement() AlterColumnBuilder
}

type ConstraintBuilder interface {
	WithName(name string) ConstraintBuilder
	WithColumns(columnNames ...string) ConstraintBuilder
	IsPrimaryKey(enablePrimary bool) ConstraintBuilder
	IsUnique(enableUnique bool) ConstraintBuilder
	IsForeignKey(enableForeign bool, referenceTableName string, referenceColumnNames ...string) ConstraintBuilder
	IsCheck(expression string) ConstraintBuilder
	Build() ConstraintDefinition
}

type IndexBuilder interface {
	WithName(name string) IndexBuilder
	WithColumns(columnNames ...string) IndexBuilder
	IsUnique(enableUnique bool) IndexBuilder
	On(partialCondition string) IndexBuilder
	Build() IndexDefinition
}
