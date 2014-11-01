package tournament

import (
	"time"
)

// Logical database operations for a tournament
type Statements interface {
	CreateUser(name, publicKey string) error
	UserExists(name string) (bool, error)
	ListUsers() ([]string, error)
	CreateMap(name, source string) error
	GetMapSource(name string) (string, error)
	ListMaps() ([]string, error)
	CreateCommit(userName, commit string, time time.Time) error
	SchemaVersion() (string, error)
}

// An implementation of statements which uses an abstracted sql connection
type Commands struct {
	tx dbcon
}

func (c *Commands) SchemaVersion() (string, error) {
	var hasVersionTable bool
	if err := c.tx.QueryRow("select count(*) > 0 from sqlite_master where type='table' and name='schema_log'").Scan(&hasVersionTable); err != nil { return "", err }
	if !hasVersionTable { return ZeroVersion, nil }
	maxSchemaVersion := ZeroVersion
	versions, err := queryStrings(c.tx, "select version from schema_log")
	if err != nil { return "", err }
	for _, version := range versions {
		if SchemaVersionLess(maxSchemaVersion, version) {
			maxSchemaVersion = version
		}
	}
	return maxSchemaVersion, nil
}

func (c *Commands) CreateUser(name, publicKey string) error {
	_, err := c.tx.Exec("insert into user(name, public_key) values(?,?)", name, publicKey)
	return err
}

func (c *Commands) UserExists(name string) (bool, error) {
	var exists bool
	err := c.tx.QueryRow("select count(name) > 0 from user where name = ?", name).Scan(&exists)
	return exists, err
}

func queryStrings(db dbcon, query string) ([]string, error) {
	if rows, err := db.Query(query); err != nil {
		return nil, err
	} else {
		var values []string
		for rows.Next() {
			var value string
			if err2 := rows.Scan(&value); err2 != nil {
				return nil, err2
			} else {
				values = append(values, value)
			}
		}
		if values == nil {
			return []string{}, nil
		} else {
			return values, nil
		}
	}
}

func (c *Commands) ListUsers() ([]string, error) {
	users, err := queryStrings(c.tx, "select name from user")
	return users, err
}

func (c *Commands) CreateMap(name, source string) error {
	_, err := c.tx.Exec("insert into map(name, source) values (?,?)", name, source)
	return err
}

func (c *Commands) GetMapSource(name string) (string, error) {
	var source string
	err := c.tx.QueryRow("select source from map where name = ?", name).Scan(&source)
	return source, err
}

func (c *Commands) ListMaps() ([]string, error) {
	maps, err := queryStrings(c.tx, "select name from map")
	return maps, err
}

func (c *Commands) CreateCommit(playerName, commitHash string, time time.Time) error {
	_, err := c.tx.Exec("insert into submission(commitHash, name, date_created) values (?,?,?)", commitHash, playerName, time)
	return err
}


/*
func (c *Database) InitTables(config Config) error {
	_, err := c.db.Exec("create table if not exists user (name text not null primary key, public_key text not null, date_created timestamp not null default current_timestamp);")
	if err != nil { return err }
	_, err = c.db.Exec("create table if not exists revision (githash text not null primary key, name text not null, date timestamp not null default current_timestamp, is_head int not null default false);")
	if err != nil { return err }
	_, err = c.db.Exec("create table if not exists map (name text primary key);")
	if err != nil { return err }
	_, err = c.db.Exec("create table if not exists match (p1 text not null, p2 text not null, map text not null, result text not null, unique (p1, p2, map));")
	if err != nil { return err }
	_, err = c.db.Exec("create table if not exists version (v text not null unique;")
	if err != nil { return err }

	_, err = c.db.Exec("insert into version(v) values (?)", Version)
	return err
}

func (c *Database) Transaction(f func(*Transaction) error) error {
	tx, err := c.db.Begin()
	if err != nil { return err }
	err = f(&Transaction{tx})
	if err == nil {
		err = tx.Commit()
	} else {
		log.Println("rollback", err)
		e2 := tx.Rollback()
		if e2 != nil {
			err = e2
		}
	}
	return err
}

type Transaction struct {
	tx *sql.Tx
}

func (t *Transaction) AddUser(name, publicKey string) error {
	_, err := t.tx.Exec("insert into user(name, public_key) values(?,?)", name, publicKey)
	return err
}

func (t *Transaction) RemoveUser(name string) error {
	_, err := t.tx.Exec("delete from user where name = ?", name)
	return err
}

func (t *Transaction) RemoveUserMatches(name string) error {
	_, err := t.tx.Exec("delete from match where p1 = ? or p2 = ?", name, name)
	return err
}

func (t *Transaction) RemoveUserRevision(name string) error {
	_, err := t.tx.Exec("delete from revision where name = ?", name)
	return err
}

func (t *Transaction) ListUsers() (*sql.Rows, error) {
	rows, err := t.tx.Query("select name, public_key from user")
	return rows, err
}

func (t *Transaction) AddMap(name string) error {
	_, err := t.tx.Exec("insert into map(name) values(?)", name)
	return err
}

func (t *Transaction) RemoveMap(name string) error {
	_, err := t.tx.Exec("delete from map where name = ?", name)
	return err
}

func (d *Database) CountUsersWithName(name string) (int, error) {
	var count int
	err := d.db.QueryRow("SELECT count(*) FROM user WHERE name=?", name).Scan(&count)
	return count, err
}

func (t *Transaction) AddRevision(commit, name string, isHead bool) error {
	var err error
	h := 0
	if isHead {
		_, err = t.tx.Exec("update revision set is_head = 0 where name = ? and githash != ?", name, commit)
		h = 1
	}
	if err == nil {
		_, err = t.tx.Exec("insert into revision (githash, name, is_head) values(?,?,?)", commit, name, h)
	}
	return err
}

func (d *Database) ListHeadRevisions() (map[string]arena.Revision, error) {
	rows, err := d.db.Query("select * from revision where is_head != 0")
	commits := map[string]arena.Revision{}
	if err != nil { return commits, err }
	for rows.Next() {
		var revision arena.Revision
	    err = rows.Scan(&revision.GitHash, &revision.Name, &revision.Date, &revision.IsHead)
	    if err != nil { break }
	    commits[revision.Name] = revision
	}
	return commits, err
}

func (d *Database) ListRevisions() (map[string][]arena.Revision, error) {
}

func (d *Database) HasResult(r1 arena.Revision, r2 arena.Revision, mapName string) (bool, error) {
	var count int
	err := d.db.QueryRow("SELECT count(*) FROM match WHERE p1=? AND p2=? AND map=?", r1.GitHash, r2.GitHash, mapName).Scan(&count)
	return count > 0, err
}

func (d *Database) ListMaps() ([]string, error) {
	rows, err := d.db.Query("select * from map")
	if err != nil { return nil, err }
	maps := []string{}
	for rows.Next() {
		var mapName string
	    if err := rows.Scan(&mapName); err != nil { return nil, err }
	    maps = append(maps, mapName)
	}
	return maps, nil
}

func (t *Transaction) AddMatch(p1, p2, mapName string, result arena.MatchResult) error {
	_, err := t.tx.Exec("insert into match(p1, p2, map, result) values (?,?,?,?)", p1, p2, mapName, string(result))
	return err
}

func (t *Transaction) FlushMapFailures() error {
	_, err := t.tx.Exec("delete from match where result = ?", string(arena.ResultFail))
	return err
}

func (d *Database) RankedMatches() ([]Match, error) {
	rows, err := d.db.Query("select r1.name, r2.name, m.map, m.result from match m join revision r1 on m.p1 = r1.githash join revision r2 on m.p2 = r2.githash where r1.is_head != 0 and r2.is_head != 0")
	if err != nil { return nil, err }
	matches := []Match{}
	for rows.Next() {
		var match Match
		var result arena.MatchResult
	    if err := rows.Scan(&match.PlayerA, &match.PlayerB, &match.Map, &result); err != nil { 
	    	return nil, err
	    }
	    match.Result = arena.MatchResult(result)
	    matches = append(matches, match)
	}
	return matches, nil
}
*/
