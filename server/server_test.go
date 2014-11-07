package server

import (
	"io"
	"net/url"
	"bytes"
	"time"
	"encoding/json"
	"strings"
	"io/ioutil"
	"net/http"
	"testing"
	"runtime"
	"net/http/httptest"
	_ "github.com/mattn/go-sqlite3"
	"github.com/GlenKelley/battleref/tournament"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
)

const (
	SamplePublicKey = "ssh-rsa AAAA01234abcd sample@public.key.com"
	SampleCommitHash = "012345"
)

func ErrorNow(t *testing.T, arg ... interface{}) {
	t.Error(arg ... )
	trace := make([]byte, 1024)
	count := runtime.Stack(trace, false)
	t.Errorf("Stack of %d bytes: %s", count, trace)
	t.FailNow()
}

type MockArena struct{
}

func (a MockArena) RunMatch(properties arena.MatchProperties, clock func()time.Time) (time.Time, arena.MatchResult, error) {
	return clock(), arena.MatchResult{}, nil
}

type MockHost struct {
}

func (r MockHost) InitRepository(name, publicKey string) error {
	return nil
}

func (r MockHost) ForkRepository(source, fork, publicKey string) error {
	return nil
}

func (r MockHost) DeleteRepository(name string) error{
	return nil
}

func (r MockHost) RepositoryURL(name string) string {
	return name
}

func Check(t *testing.T, err error) {
	if err != nil {
		ErrorNow(t, err)
	}
}

type MockRemote struct {
}

func (r MockRemote) CheckoutRepository(repoURL string) (git.Repository, error) {
	return &MockRepository{}, nil
}

type MockRepository struct {
}

func (m MockRepository) AddFiles(files []string) error {
	return nil
}

func (m MockRepository) CommitFiles(files []string, message string) error {
	return nil
}

func (m MockRepository) Push() error {
	return nil
}

func (m MockRepository) Delete() error {
	return nil
}

func (m MockRepository) Dir() string {
	return ""
}

func (m MockRepository) Head() (string, error) {
	return "", nil
}

func (m MockRepository) Log() ([]string, error) {
	return []string{}, nil
}

type MockBootstrap struct {
}

func (m MockBootstrap) PopulateRepository(name, repoURL, category string) ([]string, error) {
	return []string{}, nil
}

func createServer(t * testing.T) (*ServerState) {
	if database, err := tournament.NewInMemoryDatabase(); err != nil {
		ErrorNow(t, err)
		return nil
	} else if err = database.MigrateSchema(); err != nil {
		ErrorNow(t, err)
		return nil
	} else {
		tournament := tournament.NewTournament(database, MockArena{}, MockBootstrap{}, MockHost{}, MockRemote{})
		properties := Properties {
			":memory:",
			"8081",
			".",
			"../arena",
			".",
		}
		return NewServer(tournament, properties)
	}

}

func sendGet(t *testing.T, server *ServerState, url string) JSONResponse {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return sendRequest(t, server, http.StatusOK, req)
}

func sendPost(t *testing.T, server *ServerState, url string, body io.Reader) JSONResponse {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return sendRequest(t, server, http.StatusOK, req)
}

func sendJSONPost(t *testing.T, server *ServerState, url string, body interface{}) JSONResponse {
	return sendJSONPostExpectStatus(t, server, http.StatusOK, url, body)
}

func sendGetExpectStatus(t *testing.T, server *ServerState, expectedCode int, url string) JSONResponse {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return sendRequest(t, server, expectedCode, req)
}

func sendJSONPostExpectStatus(t *testing.T, server *ServerState, expectedCode int, url string, body interface{}) JSONResponse {
	bs, err := json.Marshal(body)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bs))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	req.Header.Set("Content-Type", "application/json")
	return sendRequest(t, server, expectedCode, req)
}

func sendRequest(t *testing.T, server *ServerState, expectedCode int, req *http.Request) JSONResponse {
	resp := httptest.NewRecorder()
	server.HttpServer.Handler.ServeHTTP(resp, req)
	if resp.Code != expectedCode {
		t.Error("Status", resp.Code, "expected", expectedCode)
		if p, err := ioutil.ReadAll(resp.Body); err == nil {
			t.Log(string(p))
		}
		t.FailNow()
	}
        var jsonResponse JSONResponse
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		t.Error(err)
		t.FailNow()
	}
	return jsonResponse
}

func TestVersion(t *testing.T) {
	server := createServer(t)
	r := sendGet(t, server, "/version")
	if r["schemaVersion"] == "" { t.FailNow() }
	if r["sourceVersion"] == "" { t.FailNow() }
}

func TestShutdown(t *testing.T) {
	server := createServer(t)
	go server.Serve()
	//Race condition of server not starting
	time.Sleep(time.Millisecond)
	r := sendPost(t, server, "/shutdown", nil)
	if r["shutdown"] == "" { t.FailNow() }
	sendGetExpectStatus(t, server, http.StatusInternalServerError, "/shutdown")
}

func TestParseForm(t *testing.T) {
	data := url.Values{}
	data.Set("foo","x")
	data.Add("bar","y")
	req, err := http.NewRequest("POST", "/?a=b", bytes.NewBufferString(data.Encode()))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	var form struct {
		A string `json:"foo" form:"foo" validate:"required"`
		B string `json:"bar" form:"bar"`
		C string `json:"moo"`
	}
	if err = parseForm(req, &form); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if form.A != "x" { t.Error(form.A, "expected x") }
	if form.B != "y" { t.Error(form.B, "expected y") }
	if form.C != "" { t.Error(form.C, "expected ''") }
}

func TestParseFormJSON(t *testing.T) {
	req, err := http.NewRequest("POST", "/", strings.NewReader("{\"foo\":\"x\",\"bar\":\"y\"}"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	req.Header.Set("Content-Type", "application/json")
	var form struct {
		A string `json:"foo" form:"foo" validate:"required"`
		B string `json:"bar" form:"bar"`
		C string `json:"moo"`
	}
	if err = parseForm(req, &form); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if form.A != "x" { t.Error(form.A, "expected x") }
	if form.B != "y" { t.Error(form.B, "expected y") }
	if form.C != "" { t.Error(form.C, "expected ''") }
}

func TestRegisterForm(t *testing.T) {
	server := createServer(t)
	r := sendPost(t, server, "/register", strings.NewReader("name=NameFoo&public_key="+url.QueryEscape(SamplePublicKey)))
	if r["name"] != "NameFoo" { t.FailNow() }
	if r["public_key"] != SamplePublicKey { t.FailNow() }
}

func TestRegisterQuery(t *testing.T) {
	server := createServer(t)
	r := sendPost(t, server, "/register?name=NameFoo&public_key="+url.QueryEscape(SamplePublicKey), nil)
	if r["name"] != "NameFoo" { t.FailNow() }
	if r["public_key"] != SamplePublicKey { t.FailNow() }
}

func TestRegisterJSON(t *testing.T) {
	server := createServer(t)
	r := sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
	if r["name"] != "NameFoo" { t.FailNow() }
	if r["public_key"] != SamplePublicKey { t.FailNow() }
}

func compareStrings(a []interface{}, b []string) bool {
	if len(a) != len(b) { return false }
	for i, as := range a {
		if as != b[i] { return false }
	}
	return true
}

func compareStringsUnordered(a []interface{}, b []string) bool {
	if len(a) != len(b) { return false }
	c := make(map[string]int)
	for _, s := range a { c[s.(string)]++ }
	for _, s := range b { c[s]-- }
	for _, i := range c { if i > 0 { return false } }
	return true
}

func TestPlayers(t *testing.T) {
	server := createServer(t)
	if r := sendGet(t, server, "/players"); len(r["players"].([]interface{})) > 0 {
		t.Error("expected no players", r)
		t.FailNow()
	}
	sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
	if r := sendGet(t, server, "/players"); !compareStrings(r["players"].([]interface{}), []string{"NameFoo"}) {
		t.Error("expected single player NameFoo", r)
		t.FailNow()
	}
	sendJSONPost(t, server, "/register", map[string]string{"name":"NameBar","public_key":SamplePublicKey})
	if r := sendGet(t, server, "/players"); !compareStringsUnordered(r["players"].([]interface{}), []string{"NameFoo", "NameBar"}) {
		t.Error("expected two players NameFoo, NameBar", r)
		t.FailNow()
	}
}

func TestMaps(t *testing.T) {
	server := createServer(t)
	if r := sendGet(t, server, "/maps"); len(r["maps"].([]interface{})) > 0 {
		t.Error("expected no maps", r)
		t.FailNow()
	}
	sendJSONPost(t, server, "/map/create", map[string]string{"name":"NameFoo","source":"SourceFoo"})
	if r := sendGet(t, server, "/maps"); !compareStrings(r["maps"].([]interface{}), []string{"NameFoo"}) {
		t.Error("expected single player NameFoo", r)
		t.FailNow()
	}
	sendJSONPost(t, server, "/map/create", map[string]string{"name":"NameBar","source":"SourceBar"})
	if r := sendGet(t, server, "/maps"); !compareStringsUnordered(r["maps"].([]interface{}), []string{"NameFoo", "NameBar"}) {
		t.Error("expected two maps NameFoo, NameBar", r)
		t.FailNow()
	}
}

func TestSubmitHash(t *testing.T) {
	if !CommitHashRegex.MatchString(SampleCommitHash) {
		t.FailNow()
	}
}

func TestSubmit(t *testing.T) {
	server := createServer(t)
	sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
	if r := sendJSONPost(t, server, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":SampleCommitHash}); r["name"] != "NameFoo" {
		t.Error(r["name"], " expected ", "NameFoo")
		t.FailNow()
	} else if r["category"] != string(tournament.CategoryGeneral) {
		t.Error(r["category"], " expected ", tournament.CategoryGeneral)
		t.FailNow()
	} else if r["commit_hash"] != SampleCommitHash {
		t.Error(r["commit_hash"], " expected ", SampleCommitHash)
		t.FailNow()
	}
}

func TestSubmitPlayerNameError(t *testing.T) {
	server := createServer(t)
	if r := sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":SampleCommitHash}); r["error"] != "Unknown player" {
		t.Error(r, "expected 'Unknown player'")
		t.FailNow()
	}
}

func TestSubmitCommitHashError(t *testing.T) {
	server := createServer(t)
	sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
	if r:= sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":"InvalidCommitHash"}); r["error"] != "Invalid commit hash" {
		t.Error(r, "expected 'Unknown player'")
		t.FailNow()
	}
}

func TestSubmitDuplicateCommitError(t *testing.T) {
	server := createServer(t)
	sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
	sendJSONPost(t, server, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":SampleCommitHash})
	sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":SampleCommitHash})
}

func TestCommits(t *testing.T) {
	server := createServer(t)
	sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
	if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); len(r["commits"].([]interface{})) > 0 {
		t.Error("expected no commits", r)
		t.FailNow()
	}
	sendJSONPost(t, server, "/submit", map[string]string{"name":"NameFoo","category":"General","commit_hash":"abcdef"})
	if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); !compareStringsUnordered(r["commits"].([]interface{}), []string{"abcdef"}) {
		t.Error("expected single commit abcdef", r)
		t.FailNow()
	}
	sendJSONPost(t, server, "/submit", map[string]string{"name":"NameFoo","category":"General","commit_hash":"012345"})
	if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); !compareStringsUnordered(r["commits"].([]interface{}), []string{"abcdef","012345"}) {
		t.Error("expected two commits abcdef, 012345", r)
		t.FailNow()
	}
}
