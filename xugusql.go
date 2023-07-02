package xugusql

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	_ "github.com/wan-maoyuan/xugusql/drive"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
)

type Config struct {
	DriverName                    string
	ServerVersion                 string
	DSN                           string
	Conn                          gorm.ConnPool
	SkipInitializeWithVersion     bool
	DefaultStringSize             uint
	DefaultDatetimePrecision      *int
	DisableWithReturning          bool
	DisableDatetimePrecision      bool
	DontSupportRenameIndex        bool
	DontSupportRenameColumn       bool
	DontSupportForShareClause     bool
	DontSupportNullAsDefaultValue bool
}

type Dialector struct {
	*Config
}

var (
	// CreateClauses create clauses
	CreateClauses = []string{"INSERT", "VALUES", "ON CONFLICT"}
	// QueryClauses query clauses
	QueryClauses []string
	// UpdateClauses update clauses
	UpdateClauses = []string{"UPDATE", "SET", "WHERE", "ORDER BY", "LIMIT"}
	// DeleteClauses delete clauses
	DeleteClauses = []string{"DELETE", "FROM", "WHERE", "ORDER BY", "LIMIT"}

	defaultDatetimePrecision = 3
)

func Open(dsn string) gorm.Dialector {
	return &Dialector{Config: &Config{DSN: dsn}}
}

func (d Dialector) Name() string {
	return "xugusql"
}

// NowFunc return now func
func (d Dialector) NowFunc(n int) func() time.Time {
	return func() time.Time {
		round := time.Second / time.Duration(math.Pow10(n))
		return time.Now().Round(round)
	}
}

func (d Dialector) Apply(config *gorm.Config) error {
	if config.NowFunc == nil {
		if d.DefaultDatetimePrecision == nil {
			d.DefaultDatetimePrecision = &defaultDatetimePrecision
		}

		// while maintaining the readability of the code, separate the business logic from
		// the general part and leave it to the function to do it here.
		config.NowFunc = d.NowFunc(*d.DefaultDatetimePrecision)
	}

	return nil
}

func (d Dialector) Initialize(db *gorm.DB) (err error) {
	if d.DriverName == "" {
		d.DriverName = "xugusql"
	}

	if d.DefaultDatetimePrecision == nil {
		d.DefaultDatetimePrecision = &defaultDatetimePrecision
	}

	if d.Conn != nil {
		db.ConnPool = d.Conn
	} else {
		db.ConnPool, err = sql.Open(d.DriverName, d.DSN)
		if err != nil {
			return err
		}
	}

	// register callbacks
	callbackConfig := &callbacks.Config{
		CreateClauses: CreateClauses,
		QueryClauses:  QueryClauses,
		UpdateClauses: UpdateClauses,
		DeleteClauses: DeleteClauses,
	}

	callbacks.RegisterDefaultCallbacks(db, callbackConfig)

	for k, v := range d.ClauseBuilders() {
		db.ClauseBuilders[k] = v
	}
	return
}

const (
	// ClauseOnConflict for clause.ClauseBuilder ON CONFLICT key
	ClauseOnConflict = "ON CONFLICT"
	// ClauseValues for clause.ClauseBuilder VALUES key
	ClauseValues = "VALUES"
	// ClauseFor for clause.ClauseBuilder FOR key
	ClauseFor = "FOR"
)

func (d Dialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	clauseBuilders := map[string]clause.ClauseBuilder{
		ClauseOnConflict: func(c clause.Clause, builder clause.Builder) {
			onConflict, ok := c.Expression.(clause.OnConflict)
			if !ok {
				c.Build(builder)
				return
			}

			builder.WriteString("ON DUPLICATE KEY UPDATE ")
			if len(onConflict.DoUpdates) == 0 {
				if s := builder.(*gorm.Statement).Schema; s != nil {
					var column clause.Column
					onConflict.DoNothing = false

					if s.PrioritizedPrimaryField != nil {
						column = clause.Column{Name: s.PrioritizedPrimaryField.DBName}
					} else if len(s.DBNames) > 0 {
						column = clause.Column{Name: s.DBNames[0]}
					}

					if column.Name != "" {
						onConflict.DoUpdates = []clause.Assignment{{Column: column, Value: column}}
					}

					builder.(*gorm.Statement).AddClause(onConflict)
				}
			}

			for idx, assignment := range onConflict.DoUpdates {
				if idx > 0 {
					builder.WriteByte(',')
				}

				builder.WriteQuoted(assignment.Column)
				builder.WriteByte('=')
				if column, ok := assignment.Value.(clause.Column); ok && column.Table == "excluded" {
					column.Table = ""
					builder.WriteString("VALUES(")
					builder.WriteQuoted(column)
					builder.WriteByte(')')
				} else {
					builder.AddVar(builder, assignment.Value)
				}
			}
		},
		ClauseValues: func(c clause.Clause, builder clause.Builder) {
			if values, ok := c.Expression.(clause.Values); ok && len(values.Columns) == 0 {
				builder.WriteString("VALUES()")
				return
			}
			c.Build(builder)
		},
	}

	if d.Config.DontSupportForShareClause {
		clauseBuilders[ClauseFor] = func(c clause.Clause, builder clause.Builder) {
			if values, ok := c.Expression.(clause.Locking); ok && strings.EqualFold(values.Strength, "SHARE") {
				builder.WriteString("LOCK IN SHARE MODE")
				return
			}
			c.Build(builder)
		}
	}

	return clauseBuilders
}

func (d Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "DEFAULT"}
}

func (d Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:        db,
				Dialector: d,
			},
		},
	}
}

func (d Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	_ = writer.WriteByte('?')
}

func (d Dialector) QuoteTo(writer clause.Writer, str string) {
	_, _ = writer.WriteString(str)
}

type localTimeInterface interface {
	In(loc *time.Location) time.Time
}

func (d Dialector) Explain(sql string, vars ...interface{}) string {
	for i, v := range vars {
		if p, ok := v.(localTimeInterface); ok {
			func(i int, t localTimeInterface) {
				defer func() {
					recover()
				}()
				vars[i] = t.In(time.Local)
			}(i, p)
		}
	}
	return logger.ExplainSQL(sql, nil, `'`, vars...)
}

func (d Dialector) DataTypeOf(field *schema.Field) string {
	if _, found := field.TagSettings["RESTRICT"]; found {
		delete(field.TagSettings, "RESTRICT")
	}

	var sqlType string

	switch field.DataType {
	case schema.Bool, schema.Int, schema.Uint, schema.Float:
		sqlType = "INTEGER"

		switch {
		case field.DataType == schema.Float:
			sqlType = "FLOAT"
		case field.Size <= 8:
			sqlType = "SMALLINT"
		}

		if val, ok := field.TagSettings["AUTOINCREMENT"]; ok && utils.CheckTruth(val) {
			sqlType += " GENERATED BY DEFAULT AS IDENTITY"
		}
	case schema.String, "VARCHAR2":
		size := field.Size
		defaultSize := d.DefaultStringSize

		if size == 0 {
			if defaultSize > 0 {
				size = int(defaultSize)
			} else {
				hasIndex := field.TagSettings["INDEX"] != "" || field.TagSettings["UNIQUE"] != ""
				// TEXT, GEOMETRY or JSON column can't have a default value
				if field.PrimaryKey || field.HasDefaultValue || hasIndex {
					size = 191 // utf8mb4
				}
			}
		}

		if size >= 2000 {
			sqlType = "CLOB"
		} else {
			sqlType = fmt.Sprintf("VARCHAR2(%d)", size)
		}

	case schema.Time:
		sqlType = "TIMESTAMP WITH TIME ZONE"
		if field.NotNull || field.PrimaryKey {
			sqlType += " NOT NULL"
		}
	case schema.Bytes:
		sqlType = "BLOB"
	default:
		sqlType = string(field.DataType)

		if strings.EqualFold(sqlType, "text") {
			sqlType = "CLOB"
		}

		if sqlType == "" {
			panic(fmt.Sprintf("invalid sql type %s (%s) for oracle", field.FieldType.Name(), field.FieldType.String()))
		}

		notNull, _ := field.TagSettings["NOT NULL"]
		unique, _ := field.TagSettings["UNIQUE"]
		additionalType := fmt.Sprintf("%s %s", notNull, unique)
		if value, ok := field.TagSettings["DEFAULT"]; ok {
			additionalType = fmt.Sprintf("%s %s %s%s", "DEFAULT", value, additionalType, func() string {
				if value, ok := field.TagSettings["COMMENT"]; ok {
					return " COMMENT " + value
				}
				return ""
			}())
		}
		sqlType = fmt.Sprintf("%v %v", sqlType, additionalType)
	}

	return sqlType
}

func (d Dialector) SavePoint(tx *gorm.DB, name string) error {
	return tx.Exec("SAVEPOINT " + name).Error
}

func (d Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return tx.Exec("ROLLBACK TO SAVEPOINT " + name).Error
}
