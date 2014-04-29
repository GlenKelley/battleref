package main

import (
	"os"
	"log"
	"strings"
	"encoding/json"
	"net/http"
	"errors"
	"database/sql"
	"bytes"
	"os/exec"
	"flag"
	"fmt"
	"regexp"
	"sort"
	"time"
	// "math/rand"
	"io/ioutil"
	_ "github.com/mattn/go-sqlite3"
)

/**
TODO: validate public keys
TODO: upload maps
TODO: run checkouts in sandbox
*/

var (
	NameRegex = regexp.MustCompile("^[\\w\\d-]+$")
	PublicKeyRegex = regexp.MustCompile("")
	CommitRegex = regexp.MustCompile("^[0-9a-f]{5,40}$")
	WinRegex = regexp.MustCompile("\\[java\\] \\[server\\]\\s+([\\w\\d-]+)\\s+\\((A|B)\\) wins \\(round (\\d+)\\)")
	ReasonRegex = regexp.MustCompile("\\[java\\] Reason: ([^\n]+)")
)

type EventType int

const (
	EventNewCommit EventType = iota
	EventNewMap    EventType = iota
	EventStart     EventType = iota
)

type Event struct {
	Name string
	Type EventType
}

type Config struct {
	DatabaseFilename string `json:"database_filename"`
	GitoliteTemplate string `json:"gitolite_template"`
	PredefinedUsers map[string]string `json:"predefined_users"`
	ServerPort string `json:"server_port"`
	GitHostname string `json:"git_hostname"`
	ApiHostname string `json:"api_hostname"`
}

func LoadConfig(filename string) (Config, error) {
	var config Config
	bs, err := ioutil.ReadFile(filename)
	if err != nil { return config, err }
	err = json.Unmarshal(bs, &config)
	if err != nil { return config, err }
	return config, nil
}

type Database struct {
	db *sql.DB
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

func OpenDatabase(filename string) (*Database, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil { return nil, err }
	return &Database{db}, nil
}

func (c *Database) InitTables(config Config) error {
	_, err := c.db.Exec("create table if not exists user (name text not null primary key, public_key text not null, date_created timestamp not null default current_timestamp);")
	if err != nil { return err }
	_, err = c.db.Exec("create table if not exists revision (githash text not null primary key, name text not null, date timestamp not null default current_timestamp, is_head int not null default false);")
	if err != nil { return err }
	_, err = c.db.Exec("create table if not exists map (name text primary key);")
	if err != nil { return err }
	_, err = c.db.Exec("create table if not exists match (p1 text not null, p2 text not null, map text not null, result text not null, unique (p1, p2, map));")
	if err != nil { return err }

	return nil
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

func (c *Database) ListUsers() (*sql.Rows, error) {
	rows, err := c.db.Query("select name, public_key from user")
	return rows, err
}

func (c *Database) AddMap(name string) error {
	_, err := c.db.Exec("insert into map(name) values(?)", name)
	return err
}

func (c *Database) RemoveMap(name string) error {
	_, err := c.db.Exec("delete from map where name = ?", name)
	return err
}

func (c *Database) CountUsersWithName(name string) (int, error) {
	var count int
	err := c.db.QueryRow("SELECT count(*) FROM user WHERE name=?", name).Scan(&count)
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

func (c *Database) ListHeadRevisions() (map[string]Revision, error) {
	rows, err := c.db.Query("select * from revision where is_head != 0")
	commits := map[string]Revision{}
	if err != nil { return commits, err }
	for rows.Next() {
		var revision Revision
	    err = rows.Scan(&revision.GitHash, &revision.Name, &revision.Date, &revision.IsHead)
	    if err != nil { break }
	    commits[revision.Name] = revision
	}
	return commits, err
}

func (c *Database) ListRevisions() (map[string][]Revision, error) {
	rows, err := c.db.Query("select * from revision")
	commits := map[string][]Revision{}
	if err != nil { return commits, err }
	for rows.Next() {
		var revision Revision
	    err = rows.Scan(&revision.GitHash, &revision.Name, &revision.Date, &revision.IsHead)
	    if err != nil { break }
	    commits[revision.Name] = append(commits[revision.Name], revision)
	}
	return commits, err
}

func (c *Database) HasResult(r1 Revision, r2 Revision, mapName string) (bool, error) {
	var count int
	err := c.db.QueryRow("SELECT count(*) FROM match WHERE p1=? AND p2=? AND map=?", r1.GitHash, r2.GitHash, mapName).Scan(&count)
	return count > 0, err
}

func (c *Database) ListMaps() ([]string, error) {
	rows, err := c.db.Query("select * from map")
	if err != nil { return nil, err }
	maps := []string{}
	for rows.Next() {
		var mapName string
	    if err := rows.Scan(&mapName); err != nil { return nil, err }
	    maps = append(maps, mapName)
	}
	return maps, nil
}

func (c *Database) AddMatch(p1, p2, mapName string, result MatchResult) error {
	_, err := c.db.Exec("insert into match(p1, p2, map, result) values (?,?,?,?)", p1, p2, mapName, string(result))
	return err
}

func (c *Database) FlushMapFailures() error {
	_, err := c.db.Exec("delete from match where result = ?", string(ResultFail))
	return err
}

type Match struct {
	PlayerA string
	PlayerB string
	Map 	string
	Result  MatchResult
}

func (m *Match) PlayerAScore() int {
	switch m.Result {
		case ResultWinA: return 2 
		case ResultTieA: return 1
		default: return 0  
	}
}

func (m *Match) PlayerBScore() int {
	switch m.Result {
		case ResultWinB: return 2 
		case ResultTieB: return 1
		default: return 0  
	}
}

func (c *Database) RankedMatches() ([]Match, error) {
	rows, err := c.db.Query("select r1.name, r2.name, m.map, m.result from match m join revision r1 on m.p1 = r1.githash join revision r2 on m.p2 = r2.githash where r1.is_head != 0 and r2.is_head != 0")
	if err != nil { return nil, err }
	matches := []Match{}
	for rows.Next() {
		var match Match
		var result string
	    if err := rows.Scan(&match.PlayerA, &match.PlayerB, &match.Map, &result); err != nil { 
	    	return nil, err
	    }
	    match.Result = MatchResult(result)
	    matches = append(matches, match)
	}
	return matches, nil
}

type ServerState struct {
	Database *Database
	Config Config
	Handler *http.ServeMux
	Events chan Event
	Es []Event
}

func NewServer(config Config, database *Database) ServerState {
	s := ServerState{database, config, http.NewServeMux(), make(chan Event, 256), []Event{}}
	go s.Referee()
	s.HandleFunc("/version", version)
	s.HandleFunc("/register", register)
	s.HandleFunc("/register/check", registerCheck)
	s.HandleFunc("/commits", commits)
	s.HandleFunc("/leaderboard", leaderboard)
	s.HandleFunc("/revision/submit", revisionSubmit)
	s.HandleFunc("/account/remove", accountRemove)
	s.HandleFunc("/map/submit", mapSubmit)
	s.HandleFunc("/map/remove", mapRemove)
	s.HandleFunc("/maps", maps)
	s.HandleFunc("/events", events)
	s.HandleFunc("/clean", clean)
	s.HandleFunc("/restart", restart)
	s.HandleFunc("/tournament/start", tournamentStart)
	return s 
}

func (s *ServerState) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request, *ServerState)) {
	s.Handler.HandleFunc(pattern, func (w http.ResponseWriter, r *http.Request) {
		handler(w, r, s)
	})
}

func main() {
	var configFilename string
	var cleanDatabase bool
	flag.StringVar(&configFilename, "config", "config.json", "environment parameters for application")
	flag.BoolVar(&cleanDatabase, "clean", false, "clean database")
	flag.Parse()

	config, err := LoadConfig(configFilename)
	if err != nil { log.Fatal(err) }

	if cleanDatabase {
		os.Remove(config.DatabaseFilename)	
	}
	database, err := OpenDatabase(config.DatabaseFilename)
	if err != nil { log.Fatal(err) }

	err = database.InitTables(config)
	if err != nil { log.Fatal(err) }

	server := NewServer(config, database)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.ServerPort), server.Handler))
}

func runAndPrint(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type RegisterForm struct {
	Name string `json:"name"`
	PublicKey string `json:"public_key"`
	Silent bool `json:"silent"`
}

func (f *RegisterForm) Validate(config Config) error {
	if f.Name == "" { return errors.New("missing name") }
	if f.PublicKey == "" { return errors.New("missing public_key") }
	_, reserved := config.PredefinedUsers[f.Name]
	if reserved {
		return fmt.Errorf("the name %s is taken", f.Name)
	}
	if !NameRegex.MatchString(f.Name) { 
		return fmt.Errorf("the name %s is invalid", f.Name)
	}
	if !PublicKeyRegex.MatchString(f.PublicKey) { 
		return fmt.Errorf("the public key %s is invalid", f.PublicKey)
	}
	return nil
}

func validateUniqueness(name string, db *Database) error {
	count, err := db.CountUsersWithName(name)
	if err != nil { return errors.New("server error") }
	if count > 0 { 
		return fmt.Errorf("the name %s is taken", name) 
	}
	return nil
}

func parseRegisterForm(r *http.Request) RegisterForm {
    var form RegisterForm
    err := json.NewDecoder(r.Body).Decode(&form)
    if err != nil {
    	form.Name = r.FormValue("name")
    	form.PublicKey = r.FormValue("public_key")
    	form.Silent = r.FormValue("silent") != ""
    }
    return form
}

func registerCheck(w http.ResponseWriter, r *http.Request, s *ServerState) {
	f := parseRegisterForm(r)
    err := f.Validate(s.Config)
    if err == nil {
    	err = validateUniqueness(f.Name, s.Database)
    }
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func register(w http.ResponseWriter, r *http.Request, s *ServerState) {
    f := parseRegisterForm(r)
    err := f.Validate(s.Config)
    if err == nil {
    	err = validateUniqueness(f.Name, s.Database)
    }
    if err == nil {
		err = addUser(s.Database, f, s.Config)
    }
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("git clone git@%s:%s.git", s.Config.GitHostname, f.Name)))
	}
}

func addUser(db *Database, form RegisterForm, config Config) error {
	return db.Transaction(func (tx *Transaction) error {
		err := tx.AddUser(form.Name, form.PublicKey);
		if err != nil { return err }
		log.Printf("added user %v\n", form)
		if !form.Silent {
			if err := updateGitolite(tx, fmt.Sprintf("added user %s", form.Name), config); err != nil { return err }
			return runAndPrint(exec.Command("./initRepo", form.Name, config.GitHostname))
		}
		return nil
	})
}

func updateGitolite(tx *Transaction, message string, config Config) error {
	err := runAndPrint(exec.Command("./checkoutGitolite", config.GitHostname))
	if err != nil { return err }

	err = os.RemoveAll("gitolite-admin/keydir")
	if err != nil { return err }
	err = os.Mkdir("gitolite-admin/keydir", 0755)
	if err != nil { return err }
	b := bytes.NewBufferString(config.GitoliteTemplate)

	keys := map[string]string{}
	writeKey := func (name, publicKey string) (string, error) {
		var err error
		if alt, ok := keys[publicKey]; ok {
			name = alt
		} else {
	    	err = ioutil.WriteFile(fmt.Sprintf("gitolite-admin/keydir/%s.pub", name), []byte(publicKey + "\n"), 0644)
	    	keys[publicKey] = name
		}
		return name, err
	}
	for name, publicKey := range config.PredefinedUsers {
		_, err = writeKey(name, publicKey)
	} 

	rows, err := tx.ListUsers()
	if err != nil { return err }
	for rows.Next() {
	    var name string
	    var publicKey string
	    err = rows.Scan(&name, &publicKey)
	    if err != nil { return err }

	    log.Println(name, publicKey)
	    canonicalName, err := writeKey(name, publicKey)
	    if err != nil { return err }
	    _, err = b.WriteString(fmt.Sprintf("repo %s\n RW+ = @admins %s\n\n", name, canonicalName))
	    if err != nil { return err }
	}
	err = ioutil.WriteFile("gitolite-admin/conf/gitolite.conf", b.Bytes(), 0755)
	if err != nil { return err }

	err = runAndPrint(exec.Command("./updateGitolite", message))
	return err
}

func version(w http.ResponseWriter, r *http.Request, s *ServerState) {
	out, err := exec.Command("git","rev-parse","HEAD").Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(out))
	}
}

type PlayerPoints struct {
	Name string
	Points int
}

type ByPoints []PlayerPoints
func (a ByPoints) Len() int           { return len(a) }
func (a ByPoints) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPoints) Less(i, j int) bool { return a[i].Points > a[j].Points }

type Ladder struct {
	Total []PlayerPoints
	ByMap map[string][]PlayerPoints
	All []Match
}

func leaderboard(w http.ResponseWriter, r *http.Request, s *ServerState) {
	matches, err := s.Database.RankedMatches()
	var b []byte
	if err == nil {
		rank := map[string]int{}
		byMap := map[string]map[string]int{}
		for _, match := range matches {
			rank[match.PlayerA] += match.PlayerAScore()
			rank[match.PlayerB] += match.PlayerBScore()
			if m, ok := byMap[match.Map]; ok {
				m[match.PlayerA] += match.PlayerAScore()
				m[match.PlayerB] += match.PlayerBScore()
			} else {
				m = make(map[string]int)
				byMap[match.Map] = m
				m[match.PlayerA] += match.PlayerAScore()
				m[match.PlayerB] += match.PlayerBScore()
			}
		}
		ladder := Ladder{[]PlayerPoints{}, make(map[string][]PlayerPoints), matches}
		for name, points := range rank {
			ladder.Total = append(ladder.Total, PlayerPoints{name, points})
		}
		sort.Sort(ByPoints(ladder.Total))
		for m, v := range byMap {
			for n, p := range v {
				ladder.ByMap[m] = append(ladder.ByMap[m], PlayerPoints{n,p})
			}
			sort.Sort(ByPoints(ladder.ByMap[m]))
		}
		b, err = json.Marshal(ladder)			
	}
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}

type Revision struct {
	Name string `json:"name"`
	GitHash string `json:"git_hash"`
	Date time.Time `json:"date"`
	IsHead bool `json:"date"`
}

func commits(w http.ResponseWriter, r *http.Request, s *ServerState) {
	revisions, err := s.Database.ListRevisions()
	var b []byte
	if err == nil {
		b, err = json.Marshal(revisions)
	}
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}

type RevisionSubmitForm struct {
	GitHash string `json:"commit"`
	Repo string `json:"repo"`
	Silent bool `json:"silent"`
}

func parseRevisionSubmitForm(r *http.Request) RevisionSubmitForm {
    var form RevisionSubmitForm
    err := json.NewDecoder(r.Body).Decode(&form)
    if err != nil {
    	form.GitHash = strings.ToLower(r.FormValue("commit"))
    	form.Repo = r.FormValue("repo")
    	form.Silent = r.FormValue("silent") != ""
    }
    return form
}

func (f *RevisionSubmitForm) Validate() error {
	if f.Repo == "" { return errors.New("missing repo") }
	if f.GitHash == "" { return errors.New("missing commit") }
	if !NameRegex.MatchString(f.Repo) { 
		return fmt.Errorf("the repo %s is invalid", f.Repo)
	}
	if !CommitRegex.MatchString(f.GitHash) { 
		return fmt.Errorf("the commit %s is invalid", f.GitHash)
	}
	return nil
}

func checkRevision(gitHash, name, gitHostname string) error {
	return runAndPrint(exec.Command("./checkBranch", name, gitHash, gitHostname))
}

func revisionSubmit(w http.ResponseWriter, r *http.Request, s *ServerState) {
	submitForm := parseRevisionSubmitForm(r)
	err := submitForm.Validate()
	if err == nil {
		var count int
		count, err = s.Database.CountUsersWithName(submitForm.Repo)
		if err == nil && count == 0 { 
			err = fmt.Errorf("invalid repository %s", submitForm.Repo)
		}
	}
	if err == nil && !submitForm.Silent {
		if err = checkRevision(submitForm.GitHash, submitForm.Repo, s.Config.GitHostname); err != nil {
			err = fmt.Errorf("invalid git hash %s", submitForm.GitHash)
		}
	}
	if err == nil {
		err = s.Database.Transaction(func (tx *Transaction) error {
			return tx.AddRevision(submitForm.GitHash, submitForm.Repo, true)
		})
	}
	if err == nil {
		s.Events <- Event{submitForm.GitHash, EventNewCommit}
	}
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

type RemoveForm struct {
	Name string `json:"name"`
}

func (f *RemoveForm) Validate() error {
	if f.Name == "" { return errors.New("missing name") }
	return nil
}

func parseRemoveForm(r *http.Request) RemoveForm {
	var form RemoveForm
    err := json.NewDecoder(r.Body).Decode(&form)
    if err != nil {
    	form.Name = r.FormValue("name")
    }
    return form
}

func removeAccount(name string, config Config, db *Database) error {
	return db.Transaction(func(tx *Transaction) error {
		if err := tx.RemoveUser(name); err != nil { return err }
		if err := tx.RemoveUserRevision(name); err != nil { return err }
		if err := tx.RemoveUserMatches(name); err != nil { return err }
		return updateGitolite(tx, fmt.Sprintf("removed user %s", name), config)
	})
}

func accountRemove(w http.ResponseWriter, r *http.Request, s *ServerState) {
	form := parseRemoveForm(r)
	err := form.Validate()
	if err == nil {
		var count int
		count, err = s.Database.CountUsersWithName(form.Name)
		if err == nil && count == 0 { 
			err = fmt.Errorf("invalid repository %s", form.Name)
		}
	}
	if err == nil {
		err = removeAccount(form.Name, s.Config, s.Database)
	}
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

type MapForm struct {
	Name string `json:"name"`
}

func (f *MapForm) Validate() error {
	if f.Name == "" { return errors.New("missing name") }
	return nil
}

func parseMapForm(r *http.Request) MapForm {
	var form MapForm
    err := json.NewDecoder(r.Body).Decode(&form)
    if err != nil {
    	form.Name = r.FormValue("name")
    }
    return form
}

func mapSubmit(w http.ResponseWriter, r *http.Request, s *ServerState) {
	form := parseMapForm(r)
	err := form.Validate()
	if err == nil {
		err = s.Database.AddMap(form.Name)
	}
	if err == nil {
		s.Events <- Event{form.Name, EventNewMap}
	}
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
	}
}


func mapRemove(w http.ResponseWriter, r *http.Request, s *ServerState) {
	form := parseMapForm(r)
	err := form.Validate()
	if err == nil {
		err = s.Database.RemoveMap(form.Name)
	}
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func maps(w http.ResponseWriter, r *http.Request, s *ServerState) {
	maps, err := s.Database.ListMaps()
	var b []byte
	if err == nil {
		b, err = json.Marshal(maps)
	}
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}

func events(w http.ResponseWriter, r *http.Request, s *ServerState) {
	b, err := json.Marshal(s.Es)
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}

func (s *ServerState) Referee() {
	for _ = range s.Events {
		// s.Es = append(s.Es, e)
		maps,err := s.Database.ListMaps()
		if err != nil { log.Println(err); continue }
		revisions,err := s.Database.ListHeadRevisions()
		if err != nil { log.Println(err); continue }
		var players []string
		for p, _ := range revisions {
			players = append(players, p)
		}

		for i, p1 := range players {
			for _, p2 := range players[i+1:] {
				for _, m := range maps {
					r1 := revisions[p1]
					r2 := revisions[p2]
					if done, err := s.Database.HasResult(r1, r2, m); err != nil {
						log.Println(err)
						continue
					} else if !done {
						result := playMatch(r1, r2, m, s.Config)
						if err := s.Database.AddMatch(r1.GitHash, r2.GitHash, m, result); err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}
	// 	switch e.Type {
	// 	case EventNewCommit:
	// 		mapName := e.Name			
	// 		s.Database.ListRevisions()
	// 	case EventNewMap:
	// 		commit := e.Name
	// 		s.Database.ListRevisions()
	// 	}
	// }
}

type MatchResult string

const (
	ResultWinA MatchResult = "A"
	ResultWinB MatchResult = "B"
	ResultTieA MatchResult = "TA"
	ResultTieB MatchResult = "TB"
	ResultFail MatchResult = "F"
)

func playMatch(r1 Revision, r2 Revision, m string, config Config) MatchResult {
	cmd := exec.Command("./runMatch", r1.Name, r1.GitHash, r2.Name, r2.GitHash, m, config.GitHostname)
	b, err := cmd.CombinedOutput()
	s := string(b)
	if err != nil {
		log.Println(err)
		log.Println(s)
		return ResultFail
	}
	log.Println(s)
	var winA bool
	if m := WinRegex.FindStringSubmatch(s); m != nil {
		winA = m[2] == "A"
	}
	var reason string
	if m := ReasonRegex.FindStringSubmatch(s); m != nil {
		reason = m[1]
	}
	var result MatchResult
	switch reason {
	case "The winning team won by getting a lot of milk.":
		if winA {
			result = ResultWinA
		} else {
			result = ResultWinB
		}
	case "The winning team won on tiebreakers.":
		if winA {
			result = ResultTieA
		} else {
			result = ResultTieB
		}
	default: result = ResultFail
	}
	return result
}

func tournamentStart(w http.ResponseWriter, r *http.Request, s *ServerState) {
	s.Database.FlushMapFailures()
	s.Events <- Event{"", EventStart}
	w.WriteHeader(http.StatusOK)
}

func clean(w http.ResponseWriter, r *http.Request, s *ServerState) {
	os.Remove(s.Config.DatabaseFilename)
	w.WriteHeader(http.StatusOK)
	os.Exit(1)
}

func restart(w http.ResponseWriter, r *http.Request, s *ServerState) {
	w.WriteHeader(http.StatusOK)
	os.Exit(1)
}





