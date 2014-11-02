package tournament

import (
	"time"
)

type TournamentCategory string
const (
	CategoryGeneral = "general"
)

type Tournament struct {
	Database Database
}

func NewTournament(database Database) *Tournament {
	return &Tournament{database}
}

func (t *Tournament) UserExists(name string) (bool, error) {
	exists, err := t.Database.UserExists(name)
	return exists, err
}

func (t *Tournament) ListUsers() ([]string, error) {
	users, err := t.Database.ListUsers()
	return users, err
}

func (t *Tournament) CreateUser(name, publicKey string) error {
	return t.Database.CreateUser(name, publicKey)
}

func (t *Tournament) CreateMap(name, source string) error {
	return t.Database.CreateMap(name, source)
}

func (t *Tournament) GetMapSource(name string) (string, error) {
	source, err := t.Database.GetMapSource(name)
	return source, err
}

func (t *Tournament) ListMaps() ([]string, error) {
	users, err := t.Database.ListMaps()
	return users, err
}

func (t *Tournament) SubmitCommit(name string, category TournamentCategory, commitHash string, time time.Time) error {
	return t.Database.CreateCommit(name, category, commitHash, time)
}


/*
import (
	"os"
	"path/filepath"
	"log"
	"strings"
	"encoding/json"
	"net/http"
	"errors"
	"bytes"
	"os/exec"
	// "flag"
	"fmt"
	"regexp"
	"sort"
	// "time"
	"io/ioutil"
//	_ "github.com/mattn/go-sqlite3"
//	"github.com/GlenKelley/battleref/arena"
)
var (
	NameRegex = regexp.MustCompile("^[\\w\\d-]+$")			//tournament usernames
	PublicKeyRegex = regexp.MustCompile("")					//SSH public key TODO: this
	CommitRegex = regexp.MustCompile("^[0-9a-f]{5,40}$")	//git hash
)

type Tournament struct {
	Database *Database
	Events chan Event
	Properties Properties
	GitServer
}

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

type Properties struct {
}

func LoadPropertiesFile(filename string) (Properties, error) {
	var config Config
	bs, err := ioutil.ReadFile(filename)
	if err != nil { return config, err }
	err = json.Unmarshal(bs, &config)
	return config, err
}

func (t *Tournament) createAccount(name String, publicKey String) error {
	if exists, err := t.isAccount(name); err != nil { return err } else if exists { return errors.New("Account exists") }
	if err := db.Database.Transaction(func (tx *Transaction) error {
		if err := tx.AddUser(name, publicKey); err != nil { return err }
		if err := t.GitServer.AddPlayerRepo(name); err != nil { return err }
		return nil
	}); err != nil { return err }
	return nil
}

func (t *Tournament) createMap(name, mapContent String) error {
	if exists, err := t.isMap(name); err != nil { return err } else if exists { return errors.New("Map Exists") }
	if err := db.Database.Transaction(func tx *Transaction) error {
		if err := tx.AddMap(name, mapContent); err != nil { return err }
		return nil
	}); err != nil { return err }
	return nil
}

func (t *Tournament) submitCommit(playerName, commit String) error {
	if 
}


type RevisionSubmitForm struct {
	GitHash string `json:"commit"`
	Repo string `json:"repo"`
	Silent bool `json:"silent"`
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
		if err = checkRevision(submitForm.GitHash, submitForm.Repo, s.Config.GitHostname, s.Config); err != nil {
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
		err = s.Database.Transaction(func(t *Transaction) error{ return t.AddMap(form.Name) })
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
		err = s.Database.Transaction(func(t *Transaction) error{ return t.RemoveMap(form.Name) })
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
						result := arena.PlayMatch(r1, r2, m, "arena/runMatch", s.Config.GitHostname, "arena/battlecode2014")
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

*/

