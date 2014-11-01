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
)
// import "code.google.com/p/gomock/gomock"

func ErrorNow(t *testing.T, arg ... interface{}) {
	t.Error(arg ... )
	trace := make([]byte, 1024)
	count := runtime.Stack(trace, false)
	t.Errorf("Stack of %d bytes: %s", count, trace)
	t.FailNow()
}

func createServer(t * testing.T) (*ServerState) {
	if database, err := tournament.NewInMemoryDatabase(); err != nil {
		ErrorNow(t, err)
		return nil
	} else if err = database.MigrateSchema(); err != nil {
		ErrorNow(t, err)
		return nil
	} else {
		tournament := tournament.NewTournament(database)
		properties := Properties {
			":memory:",
			"8081",
			".",
		}
		return NewServer(tournament, properties)
	}

}

func sendGet(t *testing.T, server *ServerState, url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return sendRequest(t, server, http.StatusOK, req)
}

func sendPost(t *testing.T, server *ServerState, url string, body io.Reader) string {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return sendRequest(t, server, http.StatusOK, req)
}

func sendJSONPost(t *testing.T, server *ServerState, url string, body interface{}) string {
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
	return sendRequest(t, server, http.StatusOK, req)
}

func sendGetExpectStatus(t *testing.T, server *ServerState, expectedCode int, url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return sendRequest(t, server, expectedCode, req)
}

func sendRequest(t *testing.T, server *ServerState, expectedCode int, req *http.Request) string {
	resp := httptest.NewRecorder()
	server.HttpServer.Handler.ServeHTTP(resp, req)
	if resp.Code != expectedCode {
		t.Error("Status", resp.Code, "expected", expectedCode)
		if p, err := ioutil.ReadAll(resp.Body); err == nil {
			t.Log(string(p))
		}
		t.FailNow()
	}
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Error(err)
		t.FailNow()
		return ""
	} else {
		return string(p)
	}
}

func TestVersion(t *testing.T) {
	server := createServer(t)
	ps := sendGet(t, server, "/version")
	if !strings.Contains(ps, "schemaVersion") {
		t.FailNow()
	}
	if !strings.Contains(ps, "sourceVersion") {
		t.FailNow()
	}
}

func TestShutdown(t *testing.T) {
	server := createServer(t)
	go server.Serve()
	//Race condition of server not starting
	time.Sleep(time.Millisecond)
	ps := sendGet(t, server, "/shutdown")
	if !strings.Contains(ps, "Shutting Down") {
		t.FailNow()
	}
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

func TestRegister(t *testing.T) {
	server := createServer(t)
	ps := sendPost(t, server, "/register", strings.NewReader("name=foo&public_key=bar\n"))
	if !strings.Contains(ps, "foo") || !strings.Contains(ps, "bar") {
		t.FailNow()
	}
}

func TestRegisterJSON(t *testing.T) {
	server := createServer(t)
	ps := sendJSONPost(t, server, "/register", map[string]string{"name":"foo","public_key":"bar"})
	if !strings.Contains(ps, "foo") || !strings.Contains(ps, "bar") {
		t.FailNow()
	}
}




