package server

import (
	"io"
	"net/url"
	"os"
	"bytes"
	"time"
	"encoding/json"
	"strings"
	"io/ioutil"
	"net/http"
	"testing"
	"net/http/httptest"
	_ "github.com/mattn/go-sqlite3"
	"github.com/GlenKelley/battleref/tournament"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
	"github.com/GlenKelley/battleref/testing"
)

const (
	SamplePublicKey = "ssh-rsa AAAA01234abcd sample@public.key.com"
	SampleCommitHash = "012345"
)

func ServerTest(test * testing.T, f func(*testutil.T, *ServerState)) {
	t := (*testutil.T)(test)
	if tempDir, err := ioutil.TempDir("", "battleref_test_git_repo"); err != nil {
		t.ErrorNow(err)
	} else {
		defer os.RemoveAll(tempDir)
		gitHost := git.NewLocalDirHost(tempDir)
		dummyArena := arena.DummyArena{}
		remote := &git.TempRemote{}
		bootstrap := &arena.MinimalBootstrap{}
		if database, err := tournament.NewInMemoryDatabase(); err != nil {
			t.ErrorNow(err)
		} else if err = database.MigrateSchema(); err != nil {
			t.ErrorNow(err)
		} else {
			tournament := tournament.NewTournament(database, dummyArena, bootstrap, gitHost, remote)
			properties := Properties {
				":memory:",
				"8081",
				".",//resource path
				"../arena",//arena resource path
				tempDir,
			}
			server := NewServer(tournament, properties)
			f(t, server)
		}
	}
}

func sendGet(t *testutil.T, server *ServerState, url string) JSONResponse {
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		return sendRequest(t, server, http.StatusOK, req)
	}
}

func sendPost(t *testutil.T, server *ServerState, url string, body io.Reader) JSONResponse {
	if req, err := http.NewRequest("POST", url, body); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		return sendRequest(t, server, http.StatusOK, req)
	}
}

func sendJSONPost(t *testutil.T, server *ServerState, url string, body interface{}) JSONResponse {
	return sendJSONPostExpectStatus(t, server, http.StatusOK, url, body)
}

func sendGetExpectStatus(t *testutil.T, server *ServerState, expectedCode int, url string) JSONResponse {
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		return sendRequest(t, server, expectedCode, req)
	}
}

func sendJSONPostExpectStatus(t *testutil.T, server *ServerState, expectedCode int, url string, body interface{}) JSONResponse {
	if bs, err := json.Marshal(body); err != nil {
		t.ErrorNow(err)
		return nil
	} else if req, err := http.NewRequest("POST", url, bytes.NewReader(bs)); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		req.Header.Set("Content-Type", "application/json")
		return sendRequest(t, server, expectedCode, req)
	}
}

func sendRequest(t *testutil.T, server *ServerState, expectedCode int, req *http.Request) JSONResponse {
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
	t.CheckError(json.NewDecoder(resp.Body).Decode(&jsonResponse))
	return jsonResponse
}

func TestVersion(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendGet(t, server, "/version")
		if r["schemaVersion"] == "" { t.FailNow() }
		if r["sourceVersion"] == "" { t.FailNow() }
	})
}

func TestShutdown(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		go server.Serve()
		//Race condition of server not starting
		time.Sleep(time.Millisecond)
		r := sendPost(t, server, "/shutdown", nil)
		if r["shutdown"] == "" { t.FailNow() }
		sendGetExpectStatus(t, server, http.StatusInternalServerError, "/shutdown")
	})
}

func TestParseForm(test *testing.T) {
	t := (*testutil.T)(test)
	data := url.Values{}
	data.Set("foo","x")
	data.Add("bar","y")
	if req, err := http.NewRequest("POST", "/?a=b", bytes.NewBufferString(data.Encode())); err != nil {
		t.ErrorNow(err)
	} else {
		var form struct {
			A string `json:"foo" form:"foo" validate:"required"`
			B string `json:"bar" form:"bar"`
			C string `json:"moo"`
		}
		t.CheckError(parseForm(req, &form))
		if form.A != "x" { t.Error(form.A, "expected x") }
		if form.B != "y" { t.Error(form.B, "expected y") }
		if form.C != "" { t.Error(form.C, "expected ''") }
	}
}

func TestParseFormJSON(test *testing.T) {
	t := (*testutil.T)(test)
	if req, err := http.NewRequest("POST", "/", strings.NewReader("{\"foo\":\"x\",\"bar\":\"y\"}")); err != nil {
		t.ErrorNow(err)
	} else {
		req.Header.Set("Content-Type", "application/json")
		var form struct {
			A string `json:"foo" form:"foo" validate:"required"`
			B string `json:"bar" form:"bar"`
			C string `json:"moo"`
		}
		t.CheckError(parseForm(req, &form))
		if form.A != "x" { t.Error(form.A, "expected x") }
		if form.B != "y" { t.Error(form.B, "expected y") }
		if form.C != "" { t.Error(form.C, "expected ''") }
	}
}

func TestRegisterForm(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendPost(t, server, "/register", strings.NewReader("name=NameFoo&public_key="+url.QueryEscape(SamplePublicKey)))
		if r["name"] != "NameFoo" { t.FailNow() }
		if r["public_key"] != SamplePublicKey { t.FailNow() }
	})
}

func TestRegisterQuery(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendPost(t, server, "/register?name=NameFoo&public_key="+url.QueryEscape(SamplePublicKey), nil)
		if r["name"] != "NameFoo" { t.FailNow() }
		if r["public_key"] != SamplePublicKey { t.FailNow() }
	})
}

func TestRegisterJSON(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
		if r["name"] != "NameFoo" { t.FailNow() }
		if r["public_key"] != SamplePublicKey { t.FailNow() }
	})
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
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		if r := sendGet(t, server, "/players"); len(r["players"].([]interface{})) > 0 {
			t.ErrorNow("expected no players", r)
		}
		sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
		if r := sendGet(t, server, "/players"); !compareStrings(r["players"].([]interface{}), []string{"NameFoo"}) {
			t.ErrorNow("expected single player NameFoo", r)
		}
		sendJSONPost(t, server, "/register", map[string]string{"name":"NameBar","public_key":SamplePublicKey})
		if r := sendGet(t, server, "/players"); !compareStringsUnordered(r["players"].([]interface{}), []string{"NameFoo", "NameBar"}) {
			t.ErrorNow("expected two players NameFoo, NameBar", r)
		}
	})
}

func TestMaps(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
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
			t.ErrorNow("expected two maps NameFoo, NameBar", r)
		}
	})
}

func TestSubmitHash(test *testing.T) {
	t := (*testutil.T)(test)
	if !CommitHashRegex.MatchString(SampleCommitHash) {
		t.ErrorNowf("Expected match from <%v>", SampleCommitHash)
	}
	for _, invalidKey := range []string{"FooBar"} {
		if CommitHashRegex.MatchString(invalidKey) {
			t.ErrorNowf("Expected match from <%v>", invalidKey)
		}
	}
}

func TestSubmit(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
		if r := sendJSONPost(t, server, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":SampleCommitHash}); r["name"] != "NameFoo" {
			t.ErrorNow(r["name"], " expected ", "NameFoo")
		} else if r["category"] != string(tournament.CategoryGeneral) {
			t.ErrorNow(r["category"], " expected ", tournament.CategoryGeneral)
		} else if r["commit_hash"] != SampleCommitHash {
			t.ErrorNow(r["commit_hash"], " expected ", SampleCommitHash)
		}
	})
}

func TestSubmitPlayerNameError(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		if r := sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":SampleCommitHash}); r["error"] != "Unknown player" {
			t.ErrorNow(r, "expected 'Unknown player'")
		}
	})
}

func TestSubmitCommitHashError(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
		if r:= sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":"InvalidCommitHash"}); r["error"] != "Invalid commit hash" {
			t.ErrorNow(r, "expected 'Unknown player'")
		}
	})
}

func TestSubmitDuplicateCommitError(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
		sendJSONPost(t, server, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":SampleCommitHash})
		sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name":"NameFoo","category":tournament.CategoryGeneral,"commit_hash":SampleCommitHash})
	})
}

func TestCommits(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name":"NameFoo","public_key":SamplePublicKey})
		if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); len(r["commits"].([]interface{})) > 0 {
			t.ErrorNow("expected no commits", r)
		}
		sendJSONPost(t, server, "/submit", map[string]string{"name":"NameFoo","category":"General","commit_hash":"abcdef"})
		if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); !compareStringsUnordered(r["commits"].([]interface{}), []string{"abcdef"}) {
			t.ErrorNow("expected single commit abcdef", r)
		}
		sendJSONPost(t, server, "/submit", map[string]string{"name":"NameFoo","category":"General","commit_hash":"012345"})
		if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); !compareStringsUnordered(r["commits"].([]interface{}), []string{"abcdef","012345"}) {
			t.ErrorNow("expected two commits abcdef, 012345", r)
		}
	})
}

