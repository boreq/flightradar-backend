// Package database provides database access.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type DatabaseType int

const (
	SQLite3 DatabaseType = iota
)

// DB becomes initialized after calling Init.
var DB *sqlx.DB
var dbType DatabaseType

// ErrNoRows lets the user access sql.ErrNoRows without importing database/sql.
var ErrNoRows = sql.ErrNoRows

var createTableQueries = []string{
	createPlanesSQL,
}

var tableNames = []string{
	"planes",
}

var log = logging.GetLogger("database")

// Init connects to the specified database.
func Init(databaseType DatabaseType, params string) (err error) {
	dbType = databaseType
	switch databaseType {
	case SQLite3:
		DB, err = sqlx.Open("sqlite3", params)
		if err != nil {
			return err
		}
		break
	default:
		return errors.New("Reached the default switch case in database.Init")
	}
	DB.MapperFunc(mapperFunc)
	return nil
}

// CreateTables creates database tables.
func CreateTables() error {
	for _, query := range createTableQueries {
		log.Debugf("Running: %s", query)
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// DropTables drops all database tables used by this program.
func DropTables() error {
	for _, tableName := range tableNames {
		query := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
		log.Debugf("Running: %s", query)
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func mapperFunc(fieldName string) string {
	var result string
	for i, ch := range fieldName {
		if i > 0 && i < len(fieldName)-1 && ch > 'A' && ch < 'Z' {
			result += "_"
		}
		result += strings.ToLower(string(ch))
	}
	return result
}
