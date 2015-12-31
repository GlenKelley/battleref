package tournament

import (
	"time"
)

// Logical database operations for a tournament
type Statements interface {
	RegisterKey(publicKey string) (int64, error)
	ListKeys() (map[int64]string, error)
	PlayerKeys() (map[string]int64, error)
	CreateUser(name, publicKey string) error
	DeleteUser(name string) error
	UserExists(name string) (bool, error)
	ListUsers() ([]string, error)
	CreateMap(name, source string) error
	GetMapSource(name string) (string, error)
	ListMaps() ([]string, error)
	ListMatches() ([]Match, error)
	LatestCommits(category TournamentCategory) ([]Submission, error)
	MapExists(name string) (bool, error)
	CreateCommit(userName string, category TournamentCategory, commit string, time time.Time) error
	ListCommits(name string, category TournamentCategory) ([]string, error)
	SchemaVersion() (string, error)
	CreateMatch(category TournamentCategory, mapName string, player1, player2 Submission, created time.Time) error
	UpdateMatch(category TournamentCategory, mapName string, player1, player2 Submission, finished time.Time, result MatchResult, replay string) error
	GetMatchResult(category TournamentCategory, mapName string, player1, player2 Submission) (MatchResult, error)
	GetMatchReplay(category TournamentCategory, mapName string, player1, palyer2 Submission) (string, error)
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

func (c *Commands) RegisterKey(publicKey string) (int64, error) {
	var id int64
	if _, err := c.tx.Exec("insert or ignore into pkey(key) values (?)", publicKey); err != nil {
		return 0, err
	} else if err := c.tx.QueryRow("select id from pkey where key = ?", publicKey).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (c *Commands) ListKeys() (map[int64]string, error) {
	if rows, err := c.tx.Query("select id, key from pkey"); err != nil {
		return nil, err
	} else {
		keys := make(map[int64]string)
		for rows.Next() {
			var id int64
			var public_key string
			if err2 := rows.Scan(&id, &public_key); err2 != nil {
				return nil, err2
			} else {
				keys[id] = public_key
			}
		}
		return keys, nil
	}
}

func (c *Commands) PlayerKeys() (map[string]int64, error) {
	if rows, err := c.tx.Query("select name, public_key from user"); err != nil {
		return nil, err
	} else {
		playerKeys := make(map[string]int64)
		for rows.Next() {
			var name string
			var id int64
			if err2 := rows.Scan(&name, &id); err2 != nil {
				return nil, err2
			} else {
				playerKeys[name] = id
			}
		}
		return playerKeys, nil
	}
}

func (c *Commands) CreateUser(name, publicKey string) error {
	if id, err := c.RegisterKey(publicKey); err != nil {
		return err
	} else {
		_, err := c.tx.Exec("insert into user(name, public_key) values(?,?)", name, id)
		return err
	}
}

func (c *Commands) DeleteUser(name string) error {
	_, err := c.tx.Exec("delete from user where name = ?", name)
	return err
}

func (c *Commands) UserExists(name string) (bool, error) {
	var exists bool
	err := c.tx.QueryRow("select count(name) > 0 from user where name = ?", name).Scan(&exists)
	return exists, err
}

func queryStrings(db dbcon, query string, args ... interface{}) ([]string, error) {
	if rows, err := db.Query(query, args ...); err != nil {
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

func (c *Commands) MapExists(name string) (bool, error) {
	var exists bool
	err := c.tx.QueryRow("select count(name) > 0 from map where name = ?", name).Scan(&exists)
	return exists, err
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

func (c *Commands) ListMatches() ([]Match, error) {
	if rows, err := c.tx.Query("select id, player1, player2, commit1, commit2, map, category, result, updated from match"); err != nil {
		return nil, err
	} else {
		var values []Match
		for rows.Next() {
			var match Match
			var result string
			if err2 := rows.Scan(&match.Id, &match.Player1, &match.Player2, &match.Commit1, &match.Commit2, &match.Map, &match.Category, &result, &match.Time); err2 != nil {
				return nil, err2
			} else {
				match.Result = MatchResult(result)
				values = append(values, match)
			}
		}
		if values == nil {
			return []Match{}, nil
		} else {
			return values, nil
		}
	}
}

func (c *Commands) LatestCommits(category TournamentCategory) ([]Submission, error) {
	if rows, err := c.tx.Query("select s1.name, s1.commithash from submission s1 left join submission s2 on s1.name = s2.name and s1.category = s2.category and s1.date_created < s2.date_created where s2.name is null and s1.category = ?", string(category)); err != nil {
		return nil, err
	} else {
		var latestCommits []Submission
		for rows.Next() {
			var name, commit string
			if err2 := rows.Scan(&name, &commit); err2 != nil {
				return nil, err2
			} else {
				latestCommits = append(latestCommits, Submission{name, commit})
			}
		}
		return latestCommits, nil
	}
}

func (c *Commands) CreateCommit(playerName string, category TournamentCategory, commitHash string, time time.Time) error {
	_, err := c.tx.Exec("insert into submission(commitHash, name, category, date_created) values (?,?,?,?)", commitHash, playerName, string(category), time)
	return err
}

func (c *Commands) ListCommits(name string, category TournamentCategory) ([]string, error) {
	commits, err := queryStrings(c.tx, "select commitHash from submission where name = ? and category = ?", name, string(category))
	return commits, err
}

func (c *Commands) CreateMatch(category TournamentCategory, mapName string, player1, player2 Submission, created time.Time) error {
	var exists bool
	if err := c.tx.QueryRow("select count(*) > 0 from match where player1 = ? and player2 = ? and commit1 = ? and commit2 = ? and map = ? and category = ?", player1.Name, player2.Name, player1.CommitHash, player2.CommitHash, mapName, string(category)).Scan(&exists); err != nil {
		return err
	} else if exists {
		return nil
	} else {
		_, err := c.tx.Exec("insert into match(category, map, player1, player2, commit1, commit2, created, updated, result) values (?, ?, ?, ?, ?, ?, ?, ?, ?)", string(category), mapName, player1.Name, player2.Name, player1.CommitHash, player2.CommitHash, created, created, MatchResultInProgress)
		return err
	}
}

func (c *Commands) UpdateMatch(category TournamentCategory, mapName string, player1, player2 Submission, finished time.Time, result MatchResult, replay string) error {
	_, err := c.tx.Exec("update match set updated = ?, result = ?, replay = ? where category = ? and map = ? and player1 = ? and player2 = ? and commit1 = ? and commit2 = ?", finished, string(result), replay, string(category), mapName, player1.Name, player2.Name, player1.CommitHash, player2.CommitHash)
	return err
}

func (c *Commands) GetMatchResult(category TournamentCategory, mapName string, player1, player2 Submission) (MatchResult, error) {
	var result string
	err := c.tx.QueryRow("select result from match where category = ? and map = ? and player1 = ? and player2 = ? and commit1 = ? and commit2 = ?", string(category), mapName, player1.Name, player2.Name, player1.CommitHash, player2.CommitHash).Scan(&result)
	if err != nil {
		return MatchResultError, err
	} else {
		return MatchResult(result), nil
	}
}

func (c *Commands) GetMatchReplay(category TournamentCategory, mapName string, player1, player2 Submission) (string, error) {
	var replay string
	err := c.tx.QueryRow("select replay from match where category = ? and map = ? and player1 = ? and player2 = ? and commit1 = ? and commit2 = ?", string(category), mapName, player1.Name, player2.Name, player1.CommitHash, player2.CommitHash).Scan(&replay)
	return replay, err
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
