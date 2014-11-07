package tournament

import (
	"os"
	"fmt"
	"sort"
	"errors"
	"strconv"
	"regexp"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var ZeroVersion = "0.0.0"

// Common operations between a transaction and db connection
type dbcon interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// The database interface for a tournament,
// operations can be run from this object as auto-commit statements or in a transaction block
type Database interface {
	Statements
	MigrateSchema() error
	TransactionBlock(f func(Statements) error) error
}

// A database implementation which uses SQLite
type SQLiteDatabase struct {
	Commands
	conn *sql.DB
}

// Creates an sqlite database backed by a file
func OpenDatabase(filename string) (Database, error) {
	if db, err := sql.Open("sqlite3", filename); err != nil {
		return nil, err
	} else {
		return &SQLiteDatabase{Commands{db}, db}, nil
	}
}

func ResetDatabase(filename string) error {
	if filename != ":memory:" {
		return  os.Remove(filename)
	} else {
		return nil
	}
}

// Creates an sqlite database in memory
func NewInMemoryDatabase() (Database, error) {
	db, err := OpenDatabase(":memory:")
	return db, err
}

// Creates a transaction block which commits iff the function argument returns no error
func (db *SQLiteDatabase) TransactionBlock(f func(Statements) error) error {
	if tx, err := db.conn.Begin(); err != nil {
		return err
	} else {
		if err2 := f(&Commands{tx}); err2 != nil {
			if err3 := tx.Rollback(); err != nil {
				return err3
			} else {
				return err2
			}

		} else {
			if err3 := tx.Commit(); err3 != nil {
				return err3
			} else {
				return nil
			}
		}
	}
}

func SchemaVersion() string {
	versions := SchemaVersionKeys(SchemaMigrations)
	return versions[len(versions)-1]
}

var SchemaVersionRegex = regexp.MustCompile("(\\d+).(\\d+).(\\d+)")

func ParseSchemaVersion(s string) (int, int, int, error) {
	ss := SchemaVersionRegex.FindStringSubmatch(s)
	if ss == nil {
		return 0,0,0, errors.New(fmt.Sprintf("Unable to parse %s", s))
	}
	major, err := strconv.Atoi(ss[1])
	if err != nil { return 0,0,0,err }
	minor, err := strconv.Atoi(ss[2])
	if err != nil { return 0,0,0,err }
	patch, err := strconv.Atoi(ss[3])
	if err != nil { return 0,0,0,err }
	return major, minor, patch, nil
}

// BySchemaVersion implements sort.Interface for []String based on the schema version semantics.
type BySchemaVersion []string

func (a BySchemaVersion) Len() int           { return len(a) }
func (a BySchemaVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySchemaVersion) Less(i, j int) bool { return SchemaVersionLess(a[i], a[j]) }

func SchemaVersionLess(a, b string) bool {
	major1, minor1, patch1, err1 := ParseSchemaVersion(a)
	major2, minor2, patch2, err2 := ParseSchemaVersion(b)
	if err1 != nil { panic(err1) }
	if err2 != nil { panic(err2) }
	return			      major1 < major2 ||
		(major1 == major2 && (minor1 < minor2 ||
		(minor1 == minor2 &&  patch1 < patch2)))
}

func SchemaVersionKeys(schemaMigrations map[string][]string) []string {
	versions := make([]string,0,len(schemaMigrations))
	for k := range schemaMigrations {
		versions = append(versions, k)
	}
	sort.Sort(BySchemaVersion(versions))
	return versions
}

// Upgrades the database schema to the highest version defined in this file
// The lack of a version table is taken to imply a clean (pre version 0) database
func (db *SQLiteDatabase) MigrateSchema() error {
	currentVersion, err := db.SchemaVersion()
	if err != nil { return err }
	versions := SchemaVersionKeys(SchemaMigrations)
	for _, version := range versions {
		if ! SchemaVersionLess(currentVersion, version) {
			continue
		}
		migration := SchemaMigrations[version]
		if tx, err := db.conn.Begin(); err != nil {
			return err
		} else {
			for _, command := range migration {
				if _, err2 := tx.Exec(command); err2 != nil {
					tx.Rollback()
					return err2
				}
			}

			if _, err2 := tx.Exec("insert into schema_log (version) values (?)", version); err2 != nil {
				tx.Rollback()
				return err2
			}

			if err2 := tx.Commit(); err2 != nil { return err2 }
		}
	}
	return nil
}

