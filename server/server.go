package server

import (
	//"os"
	"reflect"
	//	"path/filepath"
	"log"
	"strconv"
	"strings"
	//	"encoding/json"
	"bytes"
	"compress/gzip"
	"errors"
	"net"
	"net/http"
	"net/url"
	//	"bytes"
	"os/exec"
	// "flag"
	"fmt"
	"regexp"
	//	"sort"
	"encoding/json"
	"github.com/GlenKelley/battleref/git"
	"github.com/GlenKelley/battleref/simulator"
	"github.com/GlenKelley/battleref/tournament"
	"github.com/GlenKelley/battleref/web"
	"io/ioutil"
	"path/filepath"
	"time"
)

type JSONResponse map[string]interface{}

/*
const (
	HeaderContentType = "Content-Type"
	HeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"

	ContentTypeJSON = "application/json"
	ContentTypeXML = "application/xml"
)
*/

var (
	NameRegex       = regexp.MustCompile("^[\\w\\d-]+$")     //valid tournament usernames
	CommitHashRegex = regexp.MustCompile("^[0-9a-f]{5,40}$") //git hash
)

type Route struct {
	Method  string `json:"method"`
	Pattern string `json:"pattern"`
	Help    string `json:"help,omitempty"`
}

type ServerState struct {
	Tournament *tournament.Tournament
	Properties Properties
	HttpServer *http.Server
	Listener   net.Listener
	Routes     map[string]Route
}

func NewServer(tournament *tournament.Tournament, properties Properties) *ServerState {
	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%v", properties.ServerPort),
		Handler:        http.NewServeMux(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s := ServerState{tournament, properties, httpServer, nil, make(map[string]Route)}
	s.HandleFunc("GET", "/version", version, "The code version running this server.")
	s.HandleFunc("GET", "/api", api, "API documentation.")
	s.HandleFunc("GET", "/players", players, "List all registered players.")
	s.HandleFunc("GET", "/categories", categories, "List all tournament categories.")
	s.HandleFunc("GET", "/maps", maps, "List all maps.")
	s.HandleFunc("GET", "/commits", commits, "A list of submitted commits for a player in a category.")
	s.HandleFunc("GET", "/map/source", mapSource, "")
	s.HandleFunc("POST", "/shutdown", shutdown, "Turn off the server.")
	s.HandleFunc("POST", "/register", register, "Registers a player name to a public key.")
	s.HandleFunc("POST", "/map/create", createMap, "Create a map.")
	s.HandleFunc("POST", "/submit", submit, "Register a commit for a player into a category.")
	s.HandleFunc("POST", "/match/run", runMatch, "Run a single match between two submissions.")
	s.HandleFunc("POST", "/match/run/latest", runLatestMatches, "Run matches between all recent submissions.")
	s.HandleFunc("GET", "/matches", matches, "List all matches")
	s.HandleFunc("GET", "/replay", replay, "The replay log of a single match")

	return &s

	//go s.Referee()
	//s.HandleFunc("/register", register)
	//s.HandleFunc("/register/check", registerCheck)
	//s.HandleFunc("/commits", commits)
	//s.HandleFunc("/leaderboard", leaderboard)
	//s.HandleFunc("/revision/submit", revisionSubmit)
	//s.HandleFunc("/account/remove", accountRemove)
	//s.HandleFunc("/map/submit", mapSubmit)
	//s.HandleFunc("/map/remove", mapRemove)
	//s.HandleFunc("/maps", maps)
	//s.HandleFunc("/events", events)
	//s.HandleFunc("/clean", clean)
	//s.HandleFunc("/restart", restart)
	//s.HandleFunc("/tournament/start", tournamentStart)
}

func (s *ServerState) Serve() error {
	if listener, err := net.Listen("tcp", s.HttpServer.Addr); err != nil {
		return err
	} else {
		s.Listener = listener
		return s.HttpServer.Serve(listener)
	}
}

func (s *ServerState) HandleFunc(method string, pattern string, handler func(http.ResponseWriter, *http.Request, *ServerState), help string) {
	s.Routes[pattern] = Route{method, pattern, help}
	s.HttpServer.Handler.(*http.ServeMux).HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			handler(w, r, s)
		} else if r.Method == "OPTION" {
			web.WriteCorsOptionResponse(w)
		} else {
			web.WriteJsonErrorWithCode(w, fmt.Errorf("Expected method %v not %v", method, r.Method), http.StatusMethodNotAllowed)
		}
	})
}

// Environment variables
type Properties struct {
	DatabaseURL   string            `json:"database_url"`
	ServerPort    string            `json:"server_port"`
	GitServerType string            `json:"git_server"`
	GitServerConf map[string]string `json:"git_server_conf"`
	ResourcePath  string            `json:"resource_path"`
}

func (p Properties) ArenaResourcePath() string {
	return filepath.Join(p.ResourcePath, "arena", "internal", "categories")
}

func ReadProperties(env, resourcePath string) (Properties, error) {
	propertiesFilename := filepath.Join(resourcePath, "env", fmt.Sprintf("server.%s.properties", env))
	var properties Properties
	properties.ResourcePath = resourcePath
	bs, err := ioutil.ReadFile(propertiesFilename)
	if err != nil {
		return properties, err
	}
	err = json.Unmarshal(bs, &properties)
	return properties, err
}

/*
func web.WriteJsonError(w http.ResponseWriter, err error) {
	web.WriteJsonErrorWithCode(w, err, http.StatusInternalServerError)
}

func writeXML(w http.ResponseWriter, bs []byte) {
	w.Header().Add(HeaderContentType, ContentTypeXML)
	w.Header().Add(HeaderAccessControlAllowOrigin, "*")
	if _, err := w.Write(bs); err != nil {
		log.Println("failed to send response: ", err)
	}
}

func web.WriteJsonErrorWithCode(w http.ResponseWriter, err error, status int) {
	if bs, e2 := json.Marshal(JSONResponse{"error":err.Error()}); e2 != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add(HeaderContentType, ContentTypeJSON)
		w.Header().Add(HeaderAccessControlAllowOrigin, "*")
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write(bs); err != nil {
			log.Println("Failed to send error response: ", err)
		}
	}
}

func web.WriteJson(w http.ResponseWriter, response interface{}) {
	if bs, err := json.Marshal(response); err != nil {
		web.WriteJsonError(w, err)
	} else {
		w.Header().Add(HeaderContentType, ContentTypeJSON)
		w.Header().Add(HeaderAccessControlAllowOrigin, "*")
		if _, err := w.Write(bs); err != nil {
			log.Println("Failed to send response: ", err)
		}
	}
}
*/
func GitVersion(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

func version(w http.ResponseWriter, r *http.Request, s *ServerState) {
	if schemaVersion, err := s.Tournament.Database.SchemaVersion(); err != nil {
		web.WriteJsonError(w, err)
	} else if gitVersion, err := GitVersion(s.Properties.ResourcePath); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, JSONResponse{
			"schemaVersion": schemaVersion,
			"sourceVersion": gitVersion,
		})
	}
}

func api(w http.ResponseWriter, r *http.Request, s *ServerState) {
	web.WriteJson(w, JSONResponse{"routes": s.Routes})
}

func shutdown(w http.ResponseWriter, r *http.Request, s *ServerState) {
	if s.Listener == nil {
		web.WriteJsonError(w, errors.New("Server not listening"))
	} else {
		web.WriteJson(w, JSONResponse{"message": "Shutting Down"})
		if err := s.Listener.Close(); err != nil {
			log.Print(err)
		}
		s.Listener = nil
	}
}

func parseForm(r *http.Request, form interface{}) web.Error {
	formType := reflect.TypeOf(form).Elem()
	formValue := reflect.ValueOf(form).Elem()
	contentTypes := r.Header["Content-Type"]
	werr := web.NewError(http.StatusInternalServerError, "Validation errors")
	if len(contentTypes) > 0 && contentTypes[0] == "application/json" {
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(form); err != nil {
				return web.SimpleError(err)
			}
		}
	} else {
		var postValues url.Values
		if r.Method == "POST" && r.Body != nil {
			if bs, err := ioutil.ReadAll(r.Body); err != nil {
				return web.SimpleError(err)
			} else if postValues, err = url.ParseQuery(string(bs)); err != nil {
				return web.SimpleError(err)
			}
		}
		r.ParseForm()
		for i, n := 0, formType.NumField(); i < n; i++ {
			field := formType.Field(i)
			formTag := formType.Field(i).Tag.Get("form")
			postValue := postValues.Get(formTag)
			var value string
			queryValue := r.Form.Get(formTag)
			if postValue != "" {
				value = postValue
			} else if queryValue != "" {
				value = queryValue
			}
			if value != "" {
				switch field.Type.Kind() {
				case reflect.Int64:
					{
						if v, err := strconv.Atoi(value); err != nil {
							werr.AddError(web.NewErrorItem("Invalid integer", fmt.Sprintf("Unable to parse '%v' as an integer", value), field.Name, "formfield"))
						} else {
							formValue.Field(i).SetInt(int64(v))
						}
						break
					}
				case reflect.String:
					{
						formValue.Field(i).SetString(value)
						break
					}
				default:
					return web.SimpleError(fmt.Errorf("Unexpected type %v", formType.Kind()))
				}
			}
		}
	}
	for i, n := 0, formType.NumField(); i < n; i++ {
		field := formType.Field(i)
		value := formValue.Field(i)
		validateTag := field.Tag.Get("validate")
		tags := map[string]bool{}
		for _, tag := range strings.Split(validateTag, ",") {
			tags[tag] = true
		}
		if tags["required"] && value.String() == "" {
			werr.AddError(web.NewErrorItem("Missing field", fmt.Sprintf("Missing required field %v", field.Name), field.Name, "formfield"))
		}
		if tags["required"] && tags["nonzero"] && value.Int() == 0 {
			werr.AddError(web.NewErrorItem("Missing field", fmt.Sprintf("Missing required field %v", field.Name), field.Name, "formfield"))
		}
	}
	if werr.Errors() == nil {
		return nil
	} else {
		return werr
	}
}

func register(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Name      string                        `json:"name" form:"name" validate:"required"`
		PublicKey string                        `json:"public_key" form:"public_key" validate:"required"`
		Category  tournament.TournamentCategory `json:"category" form:"category" validate:"required"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if !NameRegex.MatchString(form.Name) {
		web.WriteJsonError(w, errors.New("Invalid Name"))
	} else if match := git.PublicKeyRegex.FindStringSubmatch(strings.TrimSpace(form.PublicKey)); match == nil {
		web.WriteJsonError(w, errors.New("Invalid Public Key"))
	} else if commitHash, err := s.Tournament.CreateUser(form.Name, match[1], form.Category); err != nil {
		web.WriteJsonError(w, err)
	} else if err := s.Tournament.SubmitCommit(form.Name, form.Category, commitHash, time.Now()); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, struct {
			Name      string                        `json:"name"`
			Category  tournament.TournamentCategory `json:"category"`
			PublicKey string                        `json:"public_key"`
			RepoUrl   string                        `json:"repo_url"`
			Commit    string                        `json:"commit_hash"`
		}{
			form.Name,
			form.Category,
			form.PublicKey,
			s.Tournament.GitHost.ExternalRepositoryURL(form.Name),
			commitHash,
		})
	}
}

func players(w http.ResponseWriter, r *http.Request, s *ServerState) {
	if userNames, err := s.Tournament.ListUsers(); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, JSONResponse{"players": userNames})
	}
}

func categories(w http.ResponseWriter, r *http.Request, s *ServerState) {
	categories := s.Tournament.ListCategories()
	web.WriteJson(w, JSONResponse{"categories": categories})
}

func createMap(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Name     string                        `json:"name" form:"name" validate:"required"`
		Category tournament.TournamentCategory `json:"category" form:"category" validate:"required"`
		Source   string                        `json:"source" form:"source" validate:"required"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if !NameRegex.MatchString(form.Name) {
		web.WriteJsonError(w, errors.New("Invalid Name"))
	} else if err := s.Tournament.CreateMap(form.Name, form.Source, form.Category); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, form)
	}
}

func maps(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Category tournament.TournamentCategory `json:"category" form:"category" validate:"required"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if maps, err := s.Tournament.ListMaps(form.Category); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, JSONResponse{"maps": maps})
	}
}

func mapSource(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Name     string                        `form:"name" validate:"required"`
		Category tournament.TournamentCategory `json:"category" form:"category" validate:"required"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if source, err := s.Tournament.GetMapSource(form.Name, form.Category); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteXml(w, []byte(source))
	}
}

func submit(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Name       string                        `json:"name" form:"name" validate:"required"`
		CommitHash string                        `json:"commit_hash" form:"commit_hash" validate:"required"`
		Category   tournament.TournamentCategory `json:"category" form:"category" validate:"required"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if exists, err := s.Tournament.UserExists(form.Name); err != nil {
		web.WriteJsonError(w, err)
	} else if !exists {
		web.WriteJsonError(w, errors.New("Unknown player"))
	} else if !CommitHashRegex.MatchString(form.CommitHash) {
		web.WriteJsonError(w, errors.New("Invalid commit hash"))
	} else if err := s.Tournament.SubmitCommit(form.Name, form.Category, form.CommitHash, time.Now()); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, form)
	}
}

func commits(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Name     string                        `json:"name" form:"name" validate:"required"`
		Category tournament.TournamentCategory `json:"category" form:"category" validate:"required"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if exists, err := s.Tournament.UserExists(form.Name); err != nil {
		web.WriteJsonError(w, err)
	} else if !exists {
		web.WriteJsonError(w, errors.New("Unknown player"))
	} else if commits, err := s.Tournament.ListCommits(form.Name, form.Category); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, JSONResponse{"commits": commits})
	}
}

func matches(w http.ResponseWriter, r *http.Request, s *ServerState) {
	if matches, err := s.Tournament.ListMatches(); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, JSONResponse{"matches": matches})
	}
}

func runMatch(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Player1  string                        `json:"player1" form:"player1" validate:"required"`
		Player2  string                        `json:"player2" form:"player2" validate:"required"`
		Commit1  string                        `json:"commit1" form:"commit1" validate:"required"`
		Commit2  string                        `json:"commit2" form:"commit2" validate:"required"`
		Category tournament.TournamentCategory `json:"category" form:"category" validate:"required"`
		Map      string                        `json:"map" form:"map" validate:"required"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if exists, err := s.Tournament.UserExists(form.Player1); err != nil {
		web.WriteJsonError(w, err)
	} else if !exists {
		web.WriteJsonError(w, errors.New("Unknown player1"))
	} else if exists, err := s.Tournament.UserExists(form.Player2); err != nil {
		web.WriteJsonError(w, err)
	} else if !exists {
		web.WriteJsonError(w, errors.New("Unknown player2"))
	} else if exists, err := s.Tournament.MapExists(form.Map, form.Category); err != nil {
		web.WriteJsonError(w, err)
	} else if !exists {
		web.WriteJsonError(w, errors.New("Unknown map"))
	} else if !CommitHashRegex.MatchString(form.Commit1) {
		web.WriteJsonError(w, errors.New("Invalid commit hash 1"))
	} else if !CommitHashRegex.MatchString(form.Commit2) {
		web.WriteJsonError(w, errors.New("Invalid commit hash 2"))
	} else if id, result, err := s.Tournament.RunMatch(form.Category, form.Map, tournament.Submission{form.Player1, form.Commit1}, tournament.Submission{form.Player2, form.Commit2}, tournament.SystemClock()); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, JSONResponse{"id": id, "result": result})
	}
}

func replay(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Id int64 `json:"id" form:"id" validate:"required,nonzero"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if replay, err := s.Tournament.GetMatchReplay(form.Id); err != nil {
		web.WriteJsonError(w, err)
	} else if reader, err := gzip.NewReader(bytes.NewReader(replay)); err != nil {
		web.WriteJsonError(w, err)
	} else if unzipped, err := gzip.NewReader(reader); err != nil {
		web.WriteJsonError(w, err)
	} else if replay, err := simulator.NewReplay(unzipped); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, replay)
	}
}

func runLatestMatches(w http.ResponseWriter, r *http.Request, s *ServerState) {
	var form struct {
		Category tournament.TournamentCategory `json:"category" form:"category" validate:"required"`
	}
	if err := parseForm(r, &form); err != nil {
		web.WriteJsonWebError(w, err)
	} else if err := s.Tournament.RunLatestMatches(form.Category); err != nil {
		web.WriteJsonError(w, err)
	} else {
		web.WriteJson(w, JSONResponse{"success": true})
	}
}

//type EventType int
//
//const (
//	EventNewCommit EventType = iota
//	EventNewMap    EventType = iota
//	EventStart     EventType = iota
//)
//
//type Event struct {
//	Name string
//	Type EventType
//}
//
//type Config struct {
//	GitoliteTemplate string `json:"gitolite_template"`
//	PredefinedRepos []string `json:"predefined_repos"`
//	ServerPort string `json:"server_port"`
//	GitHostname string `json:"git_hostname"`
//}

// func (c *Config) AbsoluteAppDir() string {
// 	wd, _ := os.Getwd()
// 	return filepath.Join(wd, c.AppDir)
// }

// func (c *Config) AbsoluteSandboxDir() string {
// 	wd, _ := os.Getwd()
// 	return filepath.Join(wd, c.SandboxDir)
// }

// func (c *Config) AppScriptPath(script string) string {
// 	return filepath.Join(c.AbsoluteAppDir(), script)
// }

//func (c *Config) IsPredefinedRepo(user string) bool {
//	for _, u := range c.PredefinedRepos {
//		if user == u {
//			return true
//		}
//	}
//	return false
//}
//
//func LoadConfig(filename string) (Config, error) {
//	var config Config
//	bs, err := ioutil.ReadFile(filename)
//	if err != nil { return config, err }
//	err = json.Unmarshal(bs, &config)
//	return config, err
//}
//
//type Match struct {
//	PlayerA string
//	PlayerB string
//	Map 	string
//	Result  arena.MatchResult
//}
//
//func (m *Match) PlayerAScore() int {
//	switch m.Result {
//		case arena.ResultWinA: return 2
//		case arena.ResultTieA: return 1
//		default: return 0
//	}
//}
//
//func (m *Match) PlayerBScore() int {
//	switch m.Result {
//		case arena.ResultWinB: return 2
//		case arena.ResultTieB: return 1
//		default: return 0
//	}
//}
//
//func runScript(scriptName string, args ... string) error {
//	scriptPath := filepath.Join("server/scripts", scriptName)
//	cmd := exec.Command(scriptPath, args ... )
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//	return cmd.Run()
//}
//
//type RegisterForm struct {
//	Name string `json:"name"`
//	PublicKey string `json:"public_key"`
//}
//
//func (f *RegisterForm) Validate(config Config) error {
//	if f.Name == "" { return errors.New("missing name") }
//	if f.PublicKey == "" { return errors.New("missing public_key") }
//	reserved := config.IsPredefinedRepo(f.Name)
//	if reserved {
//		return fmt.Errorf("the name %s is taken", f.Name)
//	}
//	if !NameRegex.MatchString(f.Name) {
//		return fmt.Errorf("the name %s is invalid", f.Name)
//	}
//	if !PublicKeyRegex.MatchString(f.PublicKey) {
//		return fmt.Errorf("the public key %s is invalid", f.PublicKey)
//	}
//	return nil
//}
//
//func validateUniqueness(name string, db *Database) error {
//	count, err := db.CountUsersWithName(name)
//	if err != nil { return err }
//	if count > 0 {
//		return fmt.Errorf("the name %s is taken", name)
//	}
//	return nil
//}
//
//func parseRegisterForm(r *http.Request) RegisterForm {
//    var form RegisterForm
//    err := json.NewDecoder(r.Body).Decode(&form)
//    if err != nil {
//    	form.Name = r.FormValue("name")
//    	form.PublicKey = r.FormValue("public_key")
//    }
//    return form
//}
//
//func registerCheck(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	f := parseRegisterForm(r)
//    err := f.Validate(s.Config)
//    if err == nil {
//    	err = validateUniqueness(f.Name, s.Database)
//    }
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//	}
//}
//
//func register(w http.ResponseWriter, r *http.Request, s *ServerState) {
//    f := parseRegisterForm(r)
//    err := f.Validate(s.Config)
//    if err == nil {
//    	err = validateUniqueness(f.Name, s.Database)
//    }
//    if err == nil {
//		err = addUser(s.Database, f, s.Config)
//    }
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//		w.Write([]byte(fmt.Sprintf("git clone git@%s:%s.git", s.Config.GitHostname, f.Name)))
//	}
//}
//
//func addUser(db *Database, form RegisterForm, config Config) error {
//	return db.Transaction(func (tx *Transaction) error {
//		err := tx.AddUser(form.Name, form.PublicKey);
//		if err != nil { return err }
//		log.Printf("added user %v\n", form)
//		if err := updateGitolite(tx, fmt.Sprintf("added user %s", form.Name), config); err != nil { return err }
//		return runScript("initRepo", form.Name, config.GitHostname, "server/resources/RobotPlayer.java")
//	})
//}
//
//func updateGitolite(tx *Transaction, message string, config Config) error {
//	tmpdir, err := ioutil.TempDir("", "gitolite-admin")
//	if err != nil { return err }
//	defer os.RemoveAll(tmpdir)
//	err = runScript("checkoutGitolite", config.GitHostname, tmpdir)
//	if err != nil { return err }
//
//	keydir := filepath.Join(tmpdir, "keydir")
//	// err = os.RemoveAll(keydir)
//	// if err != nil { return err }
//	// err = os.Mkdir(keydir, 0755)
//	// if err != nil { return err }
//	b := bytes.NewBufferString(config.GitoliteTemplate)
//
//	keys := map[string]string{}
//	writeKey := func (name, publicKey string) (string, error) {
//		var err error
//		if alt, ok := keys[publicKey]; ok {
//			name = alt
//		} else {
//			publicKeyFilename := filepath.Join(keydir, fmt.Sprintf("%s.pub", name))
//	    	err = ioutil.WriteFile(publicKeyFilename, []byte(publicKey + "\n"), 0644)
//	    	keys[publicKey] = name
//		}
//		return name, err
//	}
//
//	admin := "webserver_rsa"
//	bs, err := ioutil.ReadFile(filepath.Join(keydir, admin + ".pub"))
//	if err != nil { return err }
//	keys[string(bs)] = admin
//
//	// for _, name := range config.PredefinedUsers {
//	// 	var b []byte
//	// 	b, err = ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", name + ".pub"))
//	// 	if err != nil { return err }
//	// 	_, err = writeKey(name, string(b))
//	// 	if err != nil { return err }
//	// }
//
//	rows, err := tx.ListUsers()
//	if err != nil { return err }
//	for rows.Next() {
//	    var name string
//	    var publicKey string
//	    err = rows.Scan(&name, &publicKey)
//	    if err != nil { return err }
//
//	    log.Println(name, publicKey)
//	    canonicalName, err := writeKey(name, publicKey)
//	    if err != nil { return err }
//	    _, err = b.WriteString(fmt.Sprintf("repo %s\n RW+ = @admins %s\n\n", name, canonicalName))
//	    if err != nil { return err }
//	}
//	err = ioutil.WriteFile(filepath.Join(tmpdir, "conf/gitolite.conf"), b.Bytes(), 0755)
//	if err != nil { return err }
//
//	err = runScript("updateGitolite", message, tmpdir)
//	return err
//}
//
//func version(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	out, err := exec.Command("git","rev-parse","HEAD").Output()
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.Write([]byte(out))
//	}
//}
//
//type PlayerPoints struct {
//	Name string
//	Points int
//}
//
//type ByPoints []PlayerPoints
//func (a ByPoints) Len() int           { return len(a) }
//func (a ByPoints) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
//func (a ByPoints) Less(i, j int) bool { return a[i].Points > a[j].Points }
//
//type Ladder struct {
//	Total []PlayerPoints
//	ByMap map[string][]PlayerPoints
//	All []Match
//}
//
//func leaderboard(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	matches, err := s.Database.RankedMatches()
//	var b []byte
//	if err == nil {
//		rank := map[string]int{}
//		byMap := map[string]map[string]int{}
//		for _, match := range matches {
//			rank[match.PlayerA] += match.PlayerAScore()
//			rank[match.PlayerB] += match.PlayerBScore()
//			if m, ok := byMap[match.Map]; ok {
//				m[match.PlayerA] += match.PlayerAScore()
//				m[match.PlayerB] += match.PlayerBScore()
//			} else {
//				m = make(map[string]int)
//				byMap[match.Map] = m
//				m[match.PlayerA] += match.PlayerAScore()
//				m[match.PlayerB] += match.PlayerBScore()
//			}
//		}
//		ladder := Ladder{[]PlayerPoints{}, make(map[string][]PlayerPoints), matches}
//		for name, points := range rank {
//			ladder.Total = append(ladder.Total, PlayerPoints{name, points})
//		}
//		sort.Sort(ByPoints(ladder.Total))
//		for m, v := range byMap {
//			for n, p := range v {
//				ladder.ByMap[m] = append(ladder.ByMap[m], PlayerPoints{n,p})
//			}
//			sort.Sort(ByPoints(ladder.ByMap[m]))
//		}
//		b, err = json.Marshal(ladder)
//	}
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//		w.Write(b)
//	}
//}
//
//func commits(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	revisions, err := s.Database.ListRevisions()
//	var b []byte
//	if err == nil {
//		b, err = json.Marshal(revisions)
//	}
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//		w.Write(b)
//	}
//}
//
//type RevisionSubmitForm struct {
//	GitHash string `json:"commit"`
//	Repo string `json:"repo"`
//	Silent bool `json:"silent"`
//}
//
//func parseRevisionSubmitForm(r *http.Request) RevisionSubmitForm {
//    var form RevisionSubmitForm
//    err := json.NewDecoder(r.Body).Decode(&form)
//    if err != nil {
//    	form.GitHash = strings.ToLower(r.FormValue("commit"))
//    	form.Repo = r.FormValue("repo")
//    	form.Silent = r.FormValue("silent") != ""
//    }
//    return form
//}
//
//func (f *RevisionSubmitForm) Validate() error {
//	if f.Repo == "" { return errors.New("missing repo") }
//	if f.GitHash == "" { return errors.New("missing commit") }
//	if !NameRegex.MatchString(f.Repo) {
//		return fmt.Errorf("the repo %s is invalid", f.Repo)
//	}
//	if !CommitRegex.MatchString(f.GitHash) {
//		return fmt.Errorf("the commit %s is invalid", f.GitHash)
//	}
//	return nil
//}
//
//func checkRevision(gitHash, name, gitHostname string, config Config) error {
//	return runScript("checkBranch", name, gitHash, gitHostname)
//}
//
//func revisionSubmit(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	submitForm := parseRevisionSubmitForm(r)
//	err := submitForm.Validate()
//	if err == nil {
//		var count int
//		count, err = s.Database.CountUsersWithName(submitForm.Repo)
//		if err == nil && count == 0 {
//			err = fmt.Errorf("invalid repository %s", submitForm.Repo)
//		}
//	}
//	if err == nil && !submitForm.Silent {
//		if err = checkRevision(submitForm.GitHash, submitForm.Repo, s.Config.GitHostname, s.Config); err != nil {
//			err = fmt.Errorf("invalid git hash %s", submitForm.GitHash)
//		}
//	}
//	if err == nil {
//		err = s.Database.Transaction(func (tx *Transaction) error {
//			return tx.AddRevision(submitForm.GitHash, submitForm.Repo, true)
//		})
//	}
//	if err == nil {
//		s.Events <- Event{submitForm.GitHash, EventNewCommit}
//	}
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//	}
//}
//
//type RemoveForm struct {
//	Name string `json:"name"`
//}
//
//func (f *RemoveForm) Validate() error {
//	if f.Name == "" { return errors.New("missing name") }
//	return nil
//}
//
//func parseRemoveForm(r *http.Request) RemoveForm {
//	var form RemoveForm
//    err := json.NewDecoder(r.Body).Decode(&form)
//    if err != nil {
//    	form.Name = r.FormValue("name")
//    }
//    return form
//}
//
//func removeAccount(name string, config Config, db *Database) error {
//	return db.Transaction(func(tx *Transaction) error {
//		if err := tx.RemoveUser(name); err != nil { return err }
//		if err := tx.RemoveUserRevision(name); err != nil { return err }
//		if err := tx.RemoveUserMatches(name); err != nil { return err }
//		return updateGitolite(tx, fmt.Sprintf("removed user %s", name), config)
//	})
//}
//
//func accountRemove(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	form := parseRemoveForm(r)
//	err := form.Validate()
//	if err == nil {
//		var count int
//		count, err = s.Database.CountUsersWithName(form.Name)
//		if err == nil && count == 0 {
//			err = fmt.Errorf("invalid repository %s", form.Name)
//		}
//	}
//	if err == nil {
//		err = removeAccount(form.Name, s.Config, s.Database)
//	}
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//	}
//}
//
//type MapForm struct {
//	Name string `json:"name"`
//}
//
//func (f *MapForm) Validate() error {
//	if f.Name == "" { return errors.New("missing name") }
//	return nil
//}
//
//func parseMapForm(r *http.Request) MapForm {
//	var form MapForm
//    err := json.NewDecoder(r.Body).Decode(&form)
//    if err != nil {
//    	form.Name = r.FormValue("name")
//    }
//    return form
//}
//
//func mapSubmit(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	form := parseMapForm(r)
//	err := form.Validate()
//	if err == nil {
//		err = s.Database.Transaction(func(t *Transaction) error{ return t.AddMap(form.Name) })
//	}
//	if err == nil {
//		s.Events <- Event{form.Name, EventNewMap}
//	}
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//	}
//}
//
//
//func mapRemove(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	form := parseMapForm(r)
//	err := form.Validate()
//	if err == nil {
//		err = s.Database.Transaction(func(t *Transaction) error{ return t.RemoveMap(form.Name) })
//	}
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//	}
//}
//
//func maps(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	maps, err := s.Database.ListMaps()
//	var b []byte
//	if err == nil {
//		b, err = json.Marshal(maps)
//	}
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//		w.Write(b)
//	}
//}
//
//func events(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	b, err := json.Marshal(s.Es)
//    if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(http.StatusOK)
//		w.Write(b)
//	}
//}
//
//func (s *ServerState) Referee() {
//	for _ = range s.Events {
//		// s.Es = append(s.Es, e)
//		maps,err := s.Database.ListMaps()
//		if err != nil { log.Println(err); continue }
//		revisions,err := s.Database.ListHeadRevisions()
//		if err != nil { log.Println(err); continue }
//		var players []string
//		for p, _ := range revisions {
//			players = append(players, p)
//		}
//
//		for i, p1 := range players {
//			for _, p2 := range players[i+1:] {
//				for _, m := range maps {
//					r1 := revisions[p1]
//					r2 := revisions[p2]
//					if done, err := s.Database.HasResult(r1, r2, m); err != nil {
//						log.Println(err)
//						continue
//					} else if !done {
//						result := arena.PlayMatch(r1, r2, m, "arena/runMatch", s.Config.GitHostname, "arena/battlecode2014")
//						if err := s.Database.AddMatch(r1.GitHash, r2.GitHash, m, result); err != nil {
//							log.Println(err)
//						}
//					}
//				}
//			}
//		}
//	}
//	// 	switch e.Type {
//	// 	case EventNewCommit:
//	// 		mapName := e.Name
//	// 		s.Database.ListRevisions()
//	// 	case EventNewMap:
//	// 		commit := e.Name
//	// 		s.Database.ListRevisions()
//	// 	}
//	// }
//}
//
//func tournamentStart(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	s.Database.FlushMapFailures()
//	s.Events <- Event{"", EventStart}
//	w.WriteHeader(http.StatusOK)
//}
//
//func clean(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	os.Remove(s.Config.DatabaseFilename)
//	w.WriteHeader(http.StatusOK)
//	os.Exit(1)
//}
//
//func restart(w http.ResponseWriter, r *http.Request, s *ServerState) {
//	w.WriteHeader(http.StatusOK)
//	os.Exit(1)
//}
